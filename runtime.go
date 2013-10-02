package phoenix

import (
	"code.google.com/p/goconf/conf"
	"golang.struktur.de/httputils"
	"errors"
	"log"
	"net/http"
	"time"
)

// Config provides read access to the application's configuration.
type Config interface {
	GetBool(section string, option string) (bool, error)
	GetInt(section string, option string) (int, error)
	GetString(section string, option string) (string, error)
}

// Logger provides a log-only interface to the application Logger.
//
// Presently only methods for logging at the default (debug) level
// are provided, this may change in the future.
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
}

// Metadata provides access to application information such as name and version.
type Metadata interface {
	// Name returns the the configured application name,
	// or "app" if none was set.
	Name() string

	// Version returns the configured version string, 
	// or "unreleased" if no version string was provided.
	Version() string
}

// Container provides access to system data, configuration, and
// logging.
//
// Typically subinterfaces should be used when possible.
type Container interface {
	Config
	Logger
	Metadata
}

// Runtime provides application runtime support and
// server process launch functionality.
type Runtime interface {
	Container

	// DefaultHTTPHandler specifies a handler which will be run
	// using the default HTTP server configuration.
	//
	// The results of calling this method after Start() has been
	// called are undefined.
	DefaultHTTPHandler(http.Handler)

	// Start runs all registered servers and blocks until they terminate.
	Start() error
}

type startFunc func(Runtime) error

type stopFunc func(Runtime)

type callback struct {
	start startFunc
	stop stopFunc
}

type runtime struct {
	name, version string
	*log.Logger
	*conf.ConfigFile
	callbacks []callback
	server *httputils.Server
	runFunc RunFunc
}

func newRuntime(name, version string, logger *log.Logger, configFile *conf.ConfigFile, runFunc RunFunc) *runtime {
	return &runtime{name, version, logger, configFile, make([]callback, 0), nil, runFunc}
}

func (runtime *runtime) Callback(start startFunc, stop stopFunc) {
	runtime.callbacks = append(runtime.callbacks, callback{start, stop})
}

func (runtime *runtime) OnStart(start startFunc) {
	runtime.Callback(start, func(_ Runtime) {})
}

func (runtime *runtime) OnStop(stop stopFunc) {
	runtime.Callback(func(_ Runtime) error {return nil}, stop)
}

func (runtime *runtime) Run() (err error) {
	defer func() {
		if err != nil {
			runtime.Print(err)
		}
	}()

	err = runtime.runFunc(runtime)
	return
}

func (runtime *runtime) Start() error {
	if runtime.server == nil {
		return errors.New("No HTTP server was registered")
	}

	stopCallbacks := make([]callback, 0)
	defer func() {
		for _, cb := range stopCallbacks {
			cb.stop(runtime)
		}
	}()

	for _, cb := range runtime.callbacks {
		if err := cb.start(runtime); err != nil {
			return err
		} else {
			stopCallbacks = append([]callback{cb}, stopCallbacks...)
		}
	}

	return runtime.server.ListenAndServe()
}

func (runtime *runtime) DefaultHTTPHandler(handler http.Handler) {
	listen, err := runtime.GetString("http", "listen")
	if err != nil {
		listen = "127.0.0.1:8080"
	}

	readtimeout, err := runtime.GetInt("http", "readtimeout")
	if err != nil {
		readtimeout = 10
	}

	writetimeout, err := runtime.GetInt("http", "writetimeout")
	if err != nil {
		writetimeout = 10
	}

	runtime.server = &httputils.Server{
		Server: http.Server{
			Addr:           listen,
			Handler:        handler,
			ReadTimeout:    time.Duration(readtimeout) * time.Second,
			WriteTimeout:   time.Duration(writetimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		Logger: runtime.Logger,
	}

	runtime.OnStop(func(runtime Runtime) {
		runtime.Print("Server shutdown.")
	})

	runtime.OnStart(func (r Runtime) error {
		runtime.Printf("Starting HTTP server on %s", listen)
		return nil
	})
}

func (runtime *runtime) Name() string {
	if runtime.name == "" {
		return "app"
	}
	return runtime.name
}

func (runtime *runtime) Version() string {
	if runtime.version == "" {
		return "unreleased"
	}
	return runtime.version
}
