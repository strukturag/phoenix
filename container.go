package phoenix

import (
	"log"

	"code.google.com/p/goconf/conf"
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

type container struct {
	name, version string
	*log.Logger
	*conf.ConfigFile
}

func newContainer(name, version string, logger *log.Logger, configFile *conf.ConfigFile) *container {
	return &container{
		name,
		version,
		logger,
		configFile,
	}
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
