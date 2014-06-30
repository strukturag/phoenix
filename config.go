package phoenix

import (
	"code.google.com/p/goconf/conf"
)

// Config provides read access to the application's configuration.
type Config interface {
	GetBool(section, option string) (bool, error)
	GetInt(section, option string) (int, error)
	GetFloat64(section, option string) (float64, error)
	GetString(section, option string) (string, error)
}

type config struct {
	*conf.ConfigFile
	path string
	Defaults, Overrides *conf.ConfigFile
}

func newConfig() *config {
	return &config{
		ConfigFile: conf.NewConfigFile(),
		Defaults:  conf.NewConfigFile(),
		Overrides: conf.NewConfigFile(),
	}
}

func (config *config) HasPath() bool {
	return config.path != ""
}

func (config *config) Path() string {
	return config.path
}

func (config *config) SetPath(path string) {
	config.path = path
}

func (config *config) DefaultOption(section, name, value string) {
	config.Defaults.AddOption(section, name, value)
}

func (config *config) OverrideOption(section, name, value string) {
	config.Overrides.AddOption(section, name, value)
}

func (config *config) load() (err error) {
	if config.HasPath() {
		config.ConfigFile, err = conf.ReadConfigFile(config.Path())
		if err != nil {
			return
		}
	} else {
		config.ConfigFile = conf.NewConfigFile()
	}

	for _, section := range config.Defaults.GetSections() {
		options, _ := config.Defaults.GetOptions(section)
		for _, option := range options {
			if !config.ConfigFile.HasOption(section, option) {
				value, _ := config.Defaults.GetRawString(section, option)
				config.ConfigFile.AddOption(section, option, value)
			}
		}
	}

	for _, section := range config.Overrides.GetSections() {
		options, _ := config.Overrides.GetOptions(section)
		for _, option := range options {
			value, _ := config.Overrides.GetRawString(section, option)
			config.ConfigFile.AddOption(section, option, value)
		}
	}

	return
}
