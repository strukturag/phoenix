// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
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
