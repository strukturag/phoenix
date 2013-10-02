package phoenix

import (
	"testing"
)

func TestRuntime_NameUsesTheGivenValue(t *testing.T) {
	name := "spreed-app"
	runtime := newRuntime(name, "", nil, nil, nil)
	if actualName := runtime.Name(); actualName != name {
		t.Errorf("Expected app name to be '%s', but was '%s'", name, actualName)
	}
}

func TestRuntime_NameDefaultstoAppIfUnset(t *testing.T) {
	runtime := newRuntime("", "", nil, nil, nil)
	if expected, actual := "app", runtime.Name(); expected != actual {
		t.Errorf("Expected app name to be '%s', but was '%s'", expected, actual)
	}
}

func TestRuntime_VersionUsesTheGivenValue(t *testing.T) {
	version := "0.9.4b1"
	runtime := newRuntime("", version, nil, nil, nil)
	if actualVersion := runtime.Version(); actualVersion != version {
		t.Errorf("Expected app name to be '%s', but was '%s'", version, actualVersion)
	}
}

func TestRuntime_NameDefaultstoUnreleasedIfUnset(t *testing.T) {
	runtime := newRuntime("", "", nil, nil, nil)
	if expected, actual := "unreleased", runtime.Version(); expected != actual {
		t.Errorf("Expected app version to be '%s', but was '%s'", expected, actual)
	}
}
