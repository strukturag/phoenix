// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"code.google.com/p/goconf/conf"
	"fmt"
	"io"
	"log"
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
	configPath, logPath    *string
	cpuProfile, memProfile *string
	Defaults, Overrides    *conf.ConfigFile
	currentRuntime         *runtime
}

// NewServer creates a Server instance with the given name and version string.
func NewServer(name, version string) Server {
	return &server{
		Name:      name,
		Version:   version,
		Defaults:  conf.NewConfigFile(),
		Overrides: conf.NewConfigFile(),
	}
}

func (server *server) DefaultOption(section, name, value string) Server {
	server.Defaults.AddOption(section, name, value)
	return server
}

func (server *server) OverrideOption(section, name, value string) Server {
	server.Overrides.AddOption(section, name, value)
	return server
}

func (server *server) Config(path *string) Server {
	server.configPath = path
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
		return fmt.Errorf("Server is already running")
	}

	bootLogger := server.makeLogger(os.Stderr)

	configFile, err := server.loadConfig()
	if err != nil {
		bootLogger.Printf("Failed to load configuration file: %v", err)
		return err
	}

	var logfile string
	if server.logPath == nil || *server.logPath == "" {
		logfile, err = configFile.GetString("log", "logfile")
	} else {
		logfile = *server.logPath
	}

	logwriter, err := openLogWriter(logfile)
	if err != nil {
		bootLogger.Printf("Unable to open log file: %s", err)
		return err
	}
	defer logwriter.Close()

	// Set the core logging package to log to our logwriter.
	server.setSystemLogger(logwriter)

	// And create our internal logger instance.
	logger := server.makeLogger(logwriter)

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

			logger.Printf("%v\n%s", err, stackTrace)
		}
	}()

	container := newContainer(server.Name, server.Version, logger, configFile)
	runtime := newRuntime(container, runFunc)

	if server.cpuProfile != nil && *server.cpuProfile != "" {
		runtime.OnStart(func(runtime Runtime) error {
			cpuprofilepath := path.Clean(*server.cpuProfile)
			runtime.Printf("Writing CPU profile to %s", cpuprofilepath)

			f, err := os.Create(cpuprofilepath)
			if err != nil {
				return fmt.Errorf("Failed to open CPU profile: %v", err)
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
		return fmt.Errorf("Server is not currently running")
	}

	err := server.currentRuntime.Stop()
	server.currentRuntime = nil

	return err
}

func (server *server) loadConfig() (mainConfig *conf.ConfigFile, err error) {
	if server.configPath != nil {
		mainConfig, err = conf.ReadConfigFile(*server.configPath)
		if err != nil {
			return
		}
	} else {
		mainConfig = conf.NewConfigFile()
	}

	for _, section := range server.Defaults.GetSections() {
		options, _ := server.Defaults.GetOptions(section)
		for _, option := range options {
			if !mainConfig.HasOption(section, option) {
				value, _ := server.Defaults.GetRawString(section, option)
				mainConfig.AddOption(section, option, value)
			}
		}
	}

	for _, section := range server.Overrides.GetSections() {
		options, _ := server.Overrides.GetOptions(section)
		for _, option := range options {
			value, _ := server.Overrides.GetRawString(section, option)
			mainConfig.AddOption(section, option, value)
		}
	}
	return
}

func (server *server) makeLogger(w io.Writer) *log.Logger {
	return log.New(w, server.Name+" ", log.LstdFlags)
}

func (server *server) setSystemLogger(w io.Writer) {
	log.SetOutput(w)
	log.SetPrefix(server.Name + " ")
	log.SetFlags(log.LstdFlags)
}
