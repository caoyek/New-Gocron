package main

import "testing"

func TestResolveVersion(t *testing.T) {
	version, err := resolveVersion("0.0.0-f1f52d5")
	if err != nil {
		t.Fatal(err)
	}
	if version != "v0.0.0-f1f52d5" {
		t.Fatalf("expected v0.0.0-f1f52d5, got %s", version)
	}
}

func TestResolveVersionRejectsInvalidValue(t *testing.T) {
	if _, err := resolveVersion("v2.0/latest"); err == nil {
		t.Fatal("expected invalid version error")
	}
}

func TestParseValues(t *testing.T) {
	values, err := parseValues("linux, windows", map[string]bool{"linux": true, "windows": true})
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || values[0] != "linux" || values[1] != "windows" {
		t.Fatalf("unexpected values: %#v", values)
	}
}
