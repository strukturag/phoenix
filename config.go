// Copyright 2016 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	conf "github.com/dlintw/goconf"
)

// Config provides read access to the application's configuration.
//
// GetXXXDefault methods return dflt if the named option in section has no
// value. Use HasOption to determine the status of an option thus defaulted.
type Config interface {
	HasSection(section string) bool
	GetSections() []string
	GetOptions(section string) ([]string, error)
	HasOption(section, option string) bool
	GetBool(section, option string) (bool, error)
	GetBoolDefault(section, option string, dflt bool) bool
	GetInt(section, option string) (int, error)
	GetIntDefault(section, option string, dflt int) int
	GetFloat64(section, option string) (float64, error)
	GetFloat64Default(section, option string, dflt float64) float64
	GetString(section, option string) (string, error)
	GetStringDefault(section, option, dflt string) string
}

// ConfigUpdater provides access to the applications's configuration and allows
// to update it.
//
// Update method takes a string mapping like [section][option]=value. Sections
// are automatically created as needed and existing values are overwritten.
type ConfigUpdater interface {
	Config
	Update(map[string]map[string]string) error
}

type config struct {
	*conf.ConfigFile
	path                string
	defaultPath         string
	overridePath        string
	Defaults, Overrides *conf.ConfigFile
}

func newConfig() *config {
	return &config{
		ConfigFile: conf.NewConfigFile(),
		Defaults:   conf.NewConfigFile(),
		Overrides:  conf.NewConfigFile(),
	}
}

func (config *config) GetBoolDefault(section, option string, dflt bool) bool {
	if value, err := config.GetBool(section, option); err == nil {
		return value
	}
	return dflt
}

func (config *config) GetIntDefault(section, option string, dflt int) int {
	if value, err := config.GetInt(section, option); err == nil {
		return value
	}
	return dflt
}

func (config *config) GetFloat64Default(section, option string, dflt float64) float64 {
	if value, err := config.GetFloat64(section, option); err == nil {
		return value
	}
	return dflt
}

func (config *config) GetStringDefault(section, option, dflt string) string {
	if value, err := config.GetString(section, option); err == nil {
		return value
	}
	return dflt
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

func (config *config) HasDefaultPath() bool {
	return config.defaultPath != ""
}

func (config *config) DefaultPath() string {
	return config.defaultPath
}

func (config *config) SetDefaultPath(path string) {
	config.defaultPath = path
}

func (config *config) HasOverridePath() bool {
	return config.overridePath != ""
}

func (config *config) OverridePath() string {
	return config.overridePath
}

func (config *config) SetOverridePath(path string) {
	config.overridePath = path
}

func (config *config) DefaultOption(section, name, value string) {
	config.Defaults.AddOption(section, name, value)
}

func (config *config) OverrideOption(section, name, value string) {
	config.Overrides.AddOption(section, name, value)
}

func (config *config) Update(updates map[string]map[string]string) error {
	for section, options := range updates {
		for option, value := range options {
			config.ConfigFile.AddOption(section, option, value)
		}
	}

	return nil
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
	if config.HasDefaultPath() {
		// Load defaults if a path was given.
		config.Defaults, err = conf.ReadConfigFile(config.DefaultPath())
		if err != nil {
			return
		}
	}
	if config.HasOverridePath() {
		// Load overrides if a path was given.
		config.Overrides, err = conf.ReadConfigFile(config.OverridePath())
		if err != nil {
			return
		}
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
