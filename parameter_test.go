package main

import "testing"

func TestSplitParameterQuery(t *testing.T) {
	tests := []struct {
		query        string
		name         string
		parameter    string
		hasParameter bool
	}{
		{"Urban Guru Cafe", "Urban Guru Cafe", "", false},
		{"Urban Guru Cafe ->", "Urban Guru Cafe", "", true},
		{"Urban Guru Cafe -> hello world", "Urban Guru Cafe", "hello world", true},
		{"Urban Guru Cafe ->hello", "Urban Guru Cafe", "hello", true},
		{" -> hello", "", "hello", true},
		{"foo -> bar -> baz", "foo", "bar -> baz", true},
	}

	for _, test := range tests {
		name, parameter, hasParameter := splitParameterQuery(test.query)
		if name != test.name || parameter != test.parameter || hasParameter != test.hasParameter {
			t.Fatalf("splitParameterQuery(%q) = (%q, %q, %v), want (%q, %q, %v)",
				test.query, name, parameter, hasParameter, test.name, test.parameter, test.hasParameter)
		}
	}
}

func TestFindMacro(t *testing.T) {
	macros := []KmMacro{
		{UID: "1", Name: "Random Reddit"},
		{UID: "2", Name: "Restart Home Assistant"},
	}

	if macro := findMacro(macros, "restart home assistant"); macro == nil || macro.UID != "2" {
		t.Fatalf("Expected exact case-insensitive match, got %+v", macro)
	}

	if macro := findMacro(macros, "restart"); macro == nil || macro.UID != "2" {
		t.Fatalf("Expected fuzzy match, got %+v", macro)
	}

	if macro := findMacro(macros, ""); macro != nil {
		t.Fatalf("Expected no match for empty name, got %+v", macro)
	}
}

func TestLookupMacroByUID(t *testing.T) {
	macros := []KmMacro{
		{UID: "1", Name: "Random Reddit"},
		{UID: "2", Name: "Restart Home Assistant"},
	}

	if macro := lookupMacroByUID(macros, "2"); macro == nil || macro.Name != "Restart Home Assistant" {
		t.Fatalf("Expected UID lookup to find macro, got %+v", macro)
	}

	if macro := lookupMacroByUID(macros, "missing"); macro != nil {
		t.Fatalf("Expected no match for unknown UID, got %+v", macro)
	}

	if macro := lookupMacroByUID(macros, ""); macro != nil {
		t.Fatalf("Expected no match for empty UID, got %+v", macro)
	}
}
