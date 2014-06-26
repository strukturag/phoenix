// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phoenix

import (
	"testing"
)

func Test_Container_Name_UsesTheGivenValue(t *testing.T) {
	name := "spreed-app"
	runtime := newContainer(name, "", nil, nil)
	if actualName := runtime.Name(); actualName != name {
		t.Errorf("Expected app name to be '%s', but was '%s'", name, actualName)
	}
}

func Test_Container_Name_DefaultstoAppIfUnset(t *testing.T) {
	runtime := newContainer("", "", nil, nil)
	if expected, actual := "app", runtime.Name(); expected != actual {
		t.Errorf("Expected app name to be '%s', but was '%s'", expected, actual)
	}
}

func Test_Container_Version_UsesTheGivenValue(t *testing.T) {
	version := "0.9.4b1"
	runtime := newContainer("", version, nil, nil)
	if actualVersion := runtime.Version(); actualVersion != version {
		t.Errorf("Expected app name to be '%s', but was '%s'", version, actualVersion)
	}
}

func Test_Container_Name_DefaultstoUnreleasedIfUnset(t *testing.T) {
	runtime := newContainer("", "", nil, nil)
	if expected, actual := "unreleased", runtime.Version(); expected != actual {
		t.Errorf("Expected app version to be '%s', but was '%s'", expected, actual)
	}
}
