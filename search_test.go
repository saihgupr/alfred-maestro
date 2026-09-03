package main

import (
	"testing"
)

func TestSearchMacrosRAM(t *testing.T) {
	macros := []KmMacro{
		{UID: "1", Name: "Random Reddit", Category: "Reddit"},
		{UID: "2", Name: "Random Lights", Category: "Lights"},
		{UID: "3", Name: "Reload All Command Line Entities", Category: "System"},
		{UID: "4", Name: "Restart Home Assistant", Category: "Home Assistant"},
		{UID: "5", Name: "Mimestream Quits", Category: "Email"},
		{UID: "6", Name: "Snapshot Home Assistant RAM", Category: "Proxmox"},
		{UID: "7", Name: "Delete Home Assistant RAM Snapshots", Category: "Proxmox"},
		{UID: "8", Name: "Reload All YAML", Category: "Home Assistant"},
	}

	results := SearchMacros(macros, "ram")
	if len(results) == 0 {
		t.Fatal("Expected results for 'ram', got none")
	}

	// Top result MUST be "Snapshot Home Assistant RAM" (shorter name than Delete Home Assistant RAM Snapshots)
	if results[0].Name != "Snapshot Home Assistant RAM" {
		t.Fatalf("Expected first result to be 'Snapshot Home Assistant RAM', got %q", results[0].Name)
	}

	if len(results) > 1 && results[1].Name != "Delete Home Assistant RAM Snapshots" {
		t.Fatalf("Expected second result to be 'Delete Home Assistant RAM Snapshots', got %q", results[1].Name)
	}

	// Loose non-contiguous matches like "Reload All YAML" or "Mimestream Quits" must NOT match 'ram'
	for _, res := range results {
		if res.Name == "Reload All YAML" || res.Name == "Mimestream Quits" {
			t.Fatalf("Unexpected match %q for query 'ram'", res.Name)
		}
	}
}

func TestSearchMacrosTypoAndMultiToken(t *testing.T) {
	macros := []KmMacro{
		{UID: "1", Name: "Snapshot Home Assistant RAM", Category: "Proxmox"},
		{UID: "2", Name: "Restart Home Assistant", Category: "Home Assistant"},
		{UID: "3", Name: "Random Reddit", Category: "Reddit"},
	}

	// Test user's exact typo query: "snapshot homea ssinta ram"
	results := SearchMacros(macros, "snapshot homea ssinta ram")
	if len(results) == 0 {
		t.Fatal("Expected results for typo query, got none")
	}

	if results[0].Name != "Snapshot Home Assistant RAM" {
		t.Fatalf("Expected 'Snapshot Home Assistant RAM', got %q", results[0].Name)
	}
}

func TestSearchMacrosAcronym(t *testing.T) {
	macros := []KmMacro{
		{UID: "1", Name: "Keyboard Maestro Log", Category: "Tools"},
		{UID: "2", Name: "Restart Home Assistant", Category: "Home Assistant"},
		{UID: "3", Name: "KM Logs", Category: "Tools"},
	}

	resultsKM := SearchMacros(macros, "km")
	if len(resultsKM) < 2 {
		t.Fatalf("Expected at least 2 results for 'km', got %d", len(resultsKM))
	}

	resultsHA := SearchMacros(macros, "ha")
	if len(resultsHA) == 0 || resultsHA[0].Name != "Restart Home Assistant" {
		t.Fatalf("Expected 'Restart Home Assistant' for 'ha', got %+v", resultsHA)
	}
}

func TestSearchMacrosCategory(t *testing.T) {
	macros := []KmMacro{
		{UID: "1", Name: "Snapshot Home Assistant RAM", Category: "Proxmox"},
		{UID: "2", Name: "Restart Home Assistant", Category: "Home Assistant"},
	}

	results := SearchMacros(macros, "proxmox")
	if len(results) == 0 || results[0].Name != "Snapshot Home Assistant RAM" {
		t.Fatalf("Expected 'Snapshot Home Assistant RAM' for category 'proxmox', got %+v", results)
	}
}
