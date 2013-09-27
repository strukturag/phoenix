package phoenix

import (
	"code.google.com/p/goconf/conf"
	"golang.struktur.de/httputils"
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

// Container provides access to system data, configuration, and
// logging.
//
// Typically subinterfaces should be used when possible.
type Container interface {
	Config
	Logger
}

// Runtime provides application runtime support and
// server process launch functionality.
type Runtime interface {
	Container

	// DefaultHTTPHandler specifies a handler which will be run
	// using the default HTTP server configuration.
	//
	// Errors will be returned if the configuration is incorrect
	// or potentially if binding the port fails.
	//
	// While it is not defined to do so, this call should be
	// assumed to block until otherwise specified.
	//
	// It is also an error to call this function more then once.
	DefaultHTTPHandler(http.Handler) error
}

type runtime struct {
	runner *runner
	*log.Logger
	*conf.ConfigFile
}

type startFunc func(Runtime) error

type stopFunc func(Runtime)

type callback struct {
	start startFunc
	stop stopFunc
}

type runner struct {
	runtime Runtime
	callbacks []callback
	runFunc RunFunc
}

func newRunner(logger *log.Logger, configFile *conf.ConfigFile, runFunc RunFunc) *runner {
	runner := &runner{nil, make([]callback, 0), runFunc}
	runner.runtime = newRuntime(runner, logger, configFile)
	return runner
}

func (runner *runner) Callback(start startFunc, stop stopFunc) {
	runner.callbacks = append(runner.callbacks, callback{start, stop})
}

func (runner *runner) OnStart(start startFunc) {
	runner.Callback(start, func(_ Runtime) {})
}

func (runner *runner) OnStop(stop stopFunc) {
	runner.Callback(func(_ Runtime) error {return nil}, stop)
}

func (runner *runner) Run() (err error) {
	stopCallbacks := make([]callback, 0)
	defer func() {
		for _, cb := range stopCallbacks {
			cb.stop(runner.runtime)
		}
	}()

	defer func() {
		if err != nil {
			runner.runtime.Print(err)
		}
	}()

	err = runner.runFunc(runner.runtime)
	if err != nil {
		return
	}

	for _, cb := range runner.callbacks {
		if err = cb.start(runner.runtime); err != nil {
			return
		} else {
			stopCallbacks = append([]callback{cb}, stopCallbacks...)
		}
	}
	return
}

func newRuntime(runner *runner, logger *log.Logger, configFile *conf.ConfigFile) Runtime {
	return &runtime{runner, logger, configFile}
}

func (runtime *runtime) DefaultHTTPHandler(handler http.Handler) error {
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

	server := &httputils.Server{
		Server: http.Server{
			Addr:           listen,
			Handler:        handler,
			ReadTimeout:    time.Duration(readtimeout) * time.Second,
			WriteTimeout:   time.Duration(writetimeout) * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
		Logger: runtime.Logger,
	}

	runtime.runner.OnStop(func(runtime Runtime) {
		runtime.Print("Server shutdown.")
	})

	runtime.runner.OnStart(func (r Runtime) error {
		runtime.Printf("Starting HTTP server on %s", listen)
		return server.ListenAndServe()
	})
	return nil
}
