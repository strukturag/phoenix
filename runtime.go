// Copyright 2016 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Runtime provides application runtime support and
// server process launch functionality.
type Runtime interface {
	Container

	// Service specifies a Service to be managed by this runtime.
	Service(Service)

	// DefaultHTTPHandler specifies a handler which will be run
	// using the default HTTP server configuration.
	//
	// The results of calling this method after Start() has been
	// called are undefined.
	DefaultHTTPHandler(http.Handler)

	// DefaultHTTPSHandler specifies a handler which will be run
	// using the default HTTPS server configuration.
	//
	// The results of calling this method after Start() has been
	// called are undefined.
	DefaultHTTPSHandler(http.Handler)

	// TLSConfig returns the current tls.Config used with HTTPS servers
	// If no tls.Config is set, it is created using the options provided in
	// configuration. Modifications to the tls.Config the tls.Config are
	// propagated to existing HTTPS servers.
	//
	// Results of modifying the tls.Config after Start() has been called are
	// undefined.
	TLSConfig() (*tls.Config, error)

	// SetTLSConfig applies a given tls.Config to the runtime. It
	// will be used with all HTTPS servers created after SetTLSConfig
	// was called.
	SetTLSConfig(*tls.Config)

	// Start runs all registered servers and blocks until they terminate.
	Start() error
}

type startFunc func(Runtime) error

type stopFunc func(Runtime)

type callback struct {
	start startFunc
	stop  stopFunc
}

type runtime struct {
	*serviceManager
	callbacks []callback
	tlsConfig *tls.Config
	runFunc   RunFunc
	rawLogger *log.Logger
}

func newRuntime(container *container, runFunc RunFunc) *runtime {
	runtime := &runtime{
		newServiceManager(container),
		make([]callback, 0),
		nil,
		runFunc,
		container.Logger,
	}

	return runtime
}

func (runtime *runtime) Callback(start startFunc, stop stopFunc) {
	runtime.callbacks = append(runtime.callbacks, callback{start, stop})
}

func (runtime *runtime) OnStart(start startFunc) {
	runtime.Callback(start, func(_ Runtime) {})
}

func (runtime *runtime) OnStop(stop stopFunc) {
	runtime.Callback(func(_ Runtime) error { return nil }, stop)
}

func (runtime *runtime) Run() (err error) {
	defer func() {
		if err != nil {
			runtime.Print(err)
		}
	}()

	sig := make(chan os.Signal, 3)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(sig)

	go func() {
	Loop:
		for s := range sig {
			switch s {
			case os.Interrupt, syscall.SIGTERM:
				runtime.Printf("Got signal %d, stopping all services", s)
				runtime.Stop()
				break Loop
			case syscall.SIGHUP:
				runtime.Printf("Got signal %d, reloading all services", s)
				if err := runtime.Reload(); err != nil {
					runtime.Printf("Error reloading services: %v", err)
					runtime.Stop()
				}
			}
		}
	}()

	err = runtime.runFunc(runtime)
	return
}

func (runtime *runtime) TLSConfig() (*tls.Config, error) {
	var err error
	if runtime.tlsConfig == nil {
		runtime.tlsConfig, err = loadTLSConfig(runtime, "https")
	}
	return runtime.tlsConfig, err
}

func (runtime *runtime) SetTLSConfig(tlsConfig *tls.Config) {
	runtime.tlsConfig = tlsConfig
}

func (runtime *runtime) Start() error {
	stopCallbacks := make([]callback, 0)
	defer func() {
		for _, cb := range stopCallbacks {
			cb.stop(runtime)
		}
	}()

	for _, cb := range runtime.callbacks {
		if err := cb.start(runtime); err != nil {
			return err
		}

		stopCallbacks = append([]callback{cb}, stopCallbacks...)
	}

	return runtime.serviceManager.Start()
}

func (runtime *runtime) Stop() (err error) {
	if err = runtime.serviceManager.Stop(); err != nil {
		runtime.Printf("Error stopping server: %v", err)
	}
	return
}

func (runtime *runtime) Service(service Service) {
	runtime.AddService(service)
}

func (runtime *runtime) DefaultHTTPHandler(handler http.Handler) {
	runtime.appendHTTPServices("http", handler, false)
}

func (runtime *runtime) DefaultHTTPSHandler(handler http.Handler) {
	runtime.appendHTTPServices("https", handler, true)
}

func (runtime *runtime) appendHTTPServices(section string, handler http.Handler, useTLS bool) {
	listen, err := runtime.GetString(section, "listen")
	if err != nil {
		if section != "http" {
			// Only the non-TLS default service has a default listen address.
			return
		}

		listen = "127.0.0.1:8080"
	}

	readtimeout, err := runtime.GetInt(section, "readtimeout")
	if err != nil {
		readtimeout = 10
	}

	writetimeout, err := runtime.GetInt(section, "writetimeout")
	if err != nil {
		writetimeout = 10
	}

	var tlsConfig *tls.Config
	if section == "https" {
		tlsConfig, err = runtime.TLSConfig()
		if err != nil {
			runtime.OnStart(func(r Runtime) error {
				return err
			})
			return
		}
	}

	// Loop through each listen address, seperated by space
	addresses := strings.Split(listen, " ")
	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)
		if len(addr) == 0 {
			continue
		}

		runtime.Service(newHTTPService(runtime.rawLogger, handler, addr, readtimeout, writetimeout, tlsConfig))
	}
}
