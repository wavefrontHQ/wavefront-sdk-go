package internal

import (
	"testing"
)

func TestGetSemVer(t *testing.T) {
	if sdkVersion, e := GetSemVer(""); sdkVersion != 0 || e != nil {
		t.Error("Expected sdk version to be 0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.1.0"); sdkVersion != 1.01 || e != nil {
		t.Error("Expected sdk version to be 1.0100 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.1.0-SNAPSHOT"); sdkVersion != 1.01 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.1.1"); sdkVersion != 1.0101 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.10.1"); sdkVersion != 1.1001 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.1.10"); sdkVersion != 1.011 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.0.1"); sdkVersion != 1.0001 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.0.10"); sdkVersion != 1.001 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}

	if sdkVersion, e := GetSemVer("1.10.10"); sdkVersion != 1.101 || e != nil {
		t.Error("Expected sdk version to be 0.0 but got : ", sdkVersion)
	}
}
