// Copyright 2016 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"io"
	"log"
	"log/syslog"
)

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
	ConfigUpdater
	Logger
	Metadata
}

type container struct {
	name, version string
	logwriter     io.Writer
	*log.Logger
	*config
}

func newContainer(name, version string, logPath *string, config *config) (result *container, err error) {
	if config != nil {
		if err := config.load(); err != nil {
			return nil, err
		}
	}

	var logfile string
	if logPath == nil || *logPath == "" {
		logfile, _ = config.GetString("log", "logfile")
	} else {
		logfile = *logPath
	}

	var logwriter io.Writer
	var logger *log.Logger
	if logfile == "syslog" {
		logwriter, err = syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, name)
		if err != nil {
			return nil, err
		}

		// Syslog automatically adds a the tag as prefix
		setSystemLogger("", logwriter)

		logger = log.New(logwriter, "", log.LstdFlags&^(log.Ldate|log.Ltime))
	} else {
		logwriter, err = openLogWriter(logfile)
		if err != nil {
			return nil, err
		}

		// Set the core logging package to log to our logwriter.
		setSystemLogger(name, logwriter)

		// And create our internal logger instance.
		logger = makeLogger(name, logwriter)
	}

	return &container{
		name,
		version,
		logwriter,
		logger,
		config,
	}, nil
}

func (container *container) Name() string {
	if container.name == "" {
		return "app"
	}
	return container.name
}

func (container *container) Version() string {
	if container.version == "" {
		return "unreleased"
	}
	return container.version
}

func (container *container) Close() error {
	if closer, ok := container.logwriter.(io.WriteCloser); ok {
		return closer.Close()
	}
	return nil
}
