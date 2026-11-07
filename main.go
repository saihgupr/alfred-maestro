package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/deanishe/awgo"
)

var wf *aw.Workflow

// parameterSeparator is appended to a macro name by Tab autocomplete. Anything
// typed after it is passed to the macro as its parameter.
const parameterSeparator = " ->"

// splitParameterQuery splits a query like "Macro Name -> hello" into the macro
// name and parameter. hasParameter is false when the separator isn't present.
func splitParameterQuery(query string) (name string, parameter string, hasParameter bool) {
	idx := strings.Index(query, parameterSeparator)
	if idx < 0 {
		return query, "", false
	}

	return strings.TrimSpace(query[:idx]), strings.TrimSpace(query[idx+len(parameterSeparator):]), true
}

// findMacro returns the best match for name, preferring an exact name match.
func findMacro(macros []KmMacro, name string) *KmMacro {
	if name == "" {
		return nil
	}

	for i := range macros {
		if strings.EqualFold(macros[i].Name, name) {
			return &macros[i]
		}
	}

	matches := SearchMacros(macros, name)
	if len(matches) == 0 {
		return nil
	}

	return &matches[0]
}

func run() {
	var kmMacroErr error
	reload := func() (interface{}, error) {
		macrosIndex, err := getKmMacros()

		if err != nil {
			kmMacroErr = err
		}

		var macros []KmMacro
		for _, macro := range macrosIndex {
			macros = append(macros, macro)
		}

		return macros, err
	}

	// Cache KM macros for 1 seconds
	maxCache := 1 * time.Second
	var macros []KmMacro
	err := wf.Cache.LoadOrStoreJSON("kmMacros", maxCache, reload, &macros)

	if err != nil {
		// LoadOrStoreJSON() generates a new error message
		// Therefore use kmMacroErr to get the original error message
		if kmMacroErr == nil {
			wf.Fatal(err.Error())
		} else {
			wf.Fatal(kmMacroErr.Error())
		}

		return
	}

	args := wf.Args()
	var searchQuery string
	if len(args) > 0 {
		searchQuery = args[0]
	}

	// Tab autocompletes a result to "Macro Name -> ". Anything after the
	// separator is passed to the macro as its parameter.
	if name, parameter, hasParameter := splitParameterQuery(searchQuery); hasParameter {
		macro := findMacro(macros, name)
		if macro == nil {
			wf.WarnEmpty("No macros found", "Try a different query.")
			wf.SendFeedback()
			return
		}

		item := wf.NewItem(fmt.Sprintf("Run \"%s\" with parameter", macro.Name)).
			UID(macro.UID+"-parameter").
			Arg(macro.UID, parameter).
			Autocomplete(macro.Name + parameterSeparator + " " + parameter).
			Icon(&aw.Icon{Value: "dot.png"})

		if parameter == "" {
			item.Subtitle("Type a parameter, then press Enter").Valid(false)
		} else {
			item.Subtitle("Parameter: " + parameter).Valid(true)
		}

		wf.SendFeedback()
		return
	}

	filteredMacros := SearchMacros(macros, searchQuery)

	var item *aw.Item
	for _, macro := range filteredMacros {
		item = wf.NewItem(macro.Name).
			UID(macro.UID).
			Valid(true).
			Arg(macro.UID).
			Autocomplete(macro.Name + parameterSeparator + " ").
			Icon(&aw.Icon{Value: "dot.png"})
		item.NewModifier("cmd").Subtitle("Execute with parameter...").Arg(macro.UID)

		// Ctrl: Copy Macro ID
		item.NewModifier("ctrl").Subtitle("Copy Macro ID").Arg(macro.UID)

		// Alt: Reveal in KM (Swapped from Cmd)
		item.NewModifier("alt").Subtitle("Reveal macro in Keyboard Maestro").Arg(macro.UID)

		// Shift: Copy CLI command
		cliCmd := fmt.Sprintf("/usr/local/bin/keyboardmaestro %s #%s", macro.UID, macro.Name)
		item.NewModifier("shift").Subtitle("Copy CLI command").Arg(cliCmd)

		subtitle := ""
		if macro.Hotkey != "" {
			subtitle += "Hotkey: " + macro.Hotkey
		}
		if macro.TriggerString != "" {
			if subtitle != "" {
				subtitle += " | "
			}
			subtitle += "Typed: " + macro.TriggerString
		}

		if subtitle != "" {
			item.Subtitle(subtitle)
		}
	}

	if len(filteredMacros) == 0 {
		if searchQuery == "" {
			wf.WarnEmpty("No macros found", "It seems that you haven't created any macros yet.")
		} else {
			wf.WarnEmpty("No macros found", "Try a different query.")
		}
	}

	wf.SendFeedback()
}

func init() {
	wf = aw.New()
}

func main() {
	wf.Run(run)
}
