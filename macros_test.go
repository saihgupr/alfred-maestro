package main

import (
	"testing"
)

func TestGetKmMacros(t *testing.T) {
	actualMacros, err := getKmMacros()

	assertEqual(t, err, nil)
	// Just test that we can get macros successfully
	if len(actualMacros) == 0 {
		t.Fatal("Expected to find macros, but found none")
	}
}

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%s != %s", a, b)
	}
}
