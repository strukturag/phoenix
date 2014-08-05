// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"fmt"
	"io"
	// Provide pprof support via the default servemux.
	_ "net/http/pprof"
	"os"
	"path"
	goruntime "runtime"
	"runtime/pprof"
)

// RunFunc is the completion callback for server setup.
type RunFunc func(Runtime) error

// Server provides pre-startup configuration and application boot functionality.
type Server interface {
	// DefaultOption sets the default value of the named option in the given
	// section.
	DefaultOption(section, option, value string) Server

	// OverrideOption forces the named option in the given section
	// to have the given value regardless of it's state in the
	// config file.
	OverrideOption(section, option, value string) Server

	// Config sets the path to the application's main config file.
	Config(path *string) Server

	// Log sets the path to the application's logfile. Defaults to stderr if unset.
	Log(path *string) Server

	// CpuProfile runs the application with CPU profiling enabled,
	// writing the results to path.
	CpuProfile(path *string) Server

	// MemProfile runs the application with memory profiling enabled,
	// writing the results to path.
	MemProfile(path *string) Server

	// Run initializes a Runtime instance and provides it to the runner callback,
	// returning any errors produced by the callback.
	//
	// Any errors resulting from loading the configuration or opening the log
	// will be returned without calling runner.
	Run(runner RunFunc) error

	// Stop forcibly halts the running instance.
	Stop() error
}

type server struct {
	Name, Version          string
	logPath *string
	cpuProfile, memProfile *string
	currentRuntime         *runtime
	*config
}

// NewServer creates a Server instance with the given name and version string.
func NewServer(name, version string) Server {
	return &server{
		Name:    name,
		Version: version,
		config:  newConfig(),
	}
}

func (server *server) DefaultOption(section, name, value string) Server {
	server.config.DefaultOption(section, name, value)
	return server
}

func (server *server) OverrideOption(section, name, value string) Server {
	server.config.OverrideOption(section, name, value)
	return server
}

func (server *server) Config(path *string) Server {
	server.config.SetPath(*path)
	return server
}

func (server *server) Log(path *string) Server {
	server.logPath = path
	return server
}

func (server *server) CpuProfile(path *string) Server {
	server.cpuProfile = path
	return server
}

func (server *server) MemProfile(path *string) Server {
	server.memProfile = path
	return server
}

func (server *server) Run(runFunc RunFunc) (err error) {
	if server.currentRuntime != nil {
		return fmt.Errorf("server is already running")
	}

	container, err := newContainer(server.Name, server.Version, server.logPath, server.config)
	if err != nil {
		makeLogger(server.Name, os.Stderr).Print(err)
		return err
	}
	defer container.Close()

	// Now that logging is started, install a panic handler.
	defer func() {
		if recovered := recover(); recovered != nil {
			if panicedError, ok := recovered.(error); ok {
				err = panicedError
			} else {
				err = fmt.Errorf("%v", recovered)
			}

			stackTrace := make([]byte, 1024)
			for {
				n := goruntime.Stack(stackTrace, false)
				if n < len(stackTrace) {
					stackTrace = stackTrace[0:n]
					break
				}
				stackTrace = make([]byte, len(stackTrace)*2)
			}

			container.Printf("%v\n%s", err, stackTrace)
		}
	}()

	runtime := newRuntime(container, runFunc)

	if server.cpuProfile != nil && *server.cpuProfile != "" {
		runtime.OnStart(func(runtime Runtime) error {
			cpuprofilepath := path.Clean(*server.cpuProfile)
			runtime.Printf("Writing CPU profile to %s", cpuprofilepath)

			f, err := os.Create(cpuprofilepath)
			if err != nil {
				return fmt.Errorf("failed to open CPU profile: %v", err)
			}
			return pprof.StartCPUProfile(f)
		})

		runtime.OnStop(func(_ Runtime) {
			pprof.StopCPUProfile()
		})
	}

	if server.memProfile != nil && *server.memProfile != "" {
		memprofilepath := path.Clean(*server.memProfile)
		var profileData io.WriteCloser
		runtime.OnStart(func(runtime Runtime) (err error) {
			runtime.Printf("A memory profile will be written to %s on exit.", memprofilepath)
			profileData, err = os.Create(memprofilepath)
			return
		})

		runtime.OnStop(func(runtime Runtime) {
			runtime.Printf("Writing memory profile to %s", memprofilepath)
			defer profileData.Close()
			if err := pprof.Lookup("heap").WriteTo(profileData, 0); err != nil {
				runtime.Printf("Failed to create memory profile: %v", err)
			}
		})
	}

	server.currentRuntime = runtime

	err = server.currentRuntime.Run()
	return
}

func (server *server) Stop() error {
	if server.currentRuntime == nil {
		return fmt.Errorf("server is not currently running")
	}

	err := server.currentRuntime.Stop()
	server.currentRuntime = nil

	return err
}
