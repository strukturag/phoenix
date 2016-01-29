// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"log"
	"os"
	"testing"
)

func newTestContainer(name, version string) Container {
	return &container{
		name,
		version,
		nil,
		nil,
		nil,
	}
}

func Test_Container_Name_UsesTheGivenValue(t *testing.T) {
	name := "spreed-app"
	container := newTestContainer(name, "")
	if actualName := container.Name(); actualName != name {
		t.Errorf("Expected app name to be '%s', but was '%s'", name, actualName)
	}
}

func Test_Container_Name_DefaultstoAppIfUnset(t *testing.T) {
	container := newTestContainer("", "")
	if expected, actual := "app", container.Name(); expected != actual {
		t.Errorf("Expected app name to be '%s', but was '%s'", expected, actual)
	}
}

func Test_Container_Version_UsesTheGivenValue(t *testing.T) {
	version := "0.9.4b1"
	container := newTestContainer("", version)
	if actualVersion := container.Version(); actualVersion != version {
		t.Errorf("Expected app name to be '%s', but was '%s'", version, actualVersion)
	}
}

func Test_Container_Name_DefaultstoUnreleasedIfUnset(t *testing.T) {
	container := newTestContainer("", "")
	if expected, actual := "unreleased", container.Version(); expected != actual {
		t.Errorf("Expected app version to be '%s', but was '%s'", expected, actual)
	}
}

func Test_Container_Syslog(t *testing.T) {
	logFilename := "syslog"
	container, err := newContainer("test", "", &logFilename, nil)
	if err != nil {
		t.Fatalf("Could not create container: '%s'", err)
	}
	if err := container.Close(); err != nil {
		t.Errorf("Closing the syslog container returned '%s'", err)
	}
	log.Println("Testing")
	fp, err := os.Open(logFilename)
	if fp != nil {
		fp.Close()
	}
	if err == nil {
		t.Errorf("Logging created a file '%s' but should have logged to syslog", logFilename)
	} else if !os.IsNotExist(err) {
		t.Errorf("Logfile '%s' could not be opened but should not exist: '%s'", logFilename, err)
	}
}
