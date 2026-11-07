package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/deanishe/awgo"
)

var wf *aw.Workflow

// parameterSeparator marks the parameter in a manually typed query like
// "Macro Name -> hello". (Tab entry uses session variables instead, so the
// input field holds only the parameter text.)
const parameterSeparator = " ->"

// Session variables used to remember which macro Tab was pressed on. Alfred
// passes these back in as environment variables when the script reruns.
const (
	sessionMacroUID  = "am_macro_uid"
	sessionMacroName = "am_macro_name"
)

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

// lookupMacroByUID returns the macro with the given UID, or nil.
func lookupMacroByUID(macros []KmMacro, uid string) *KmMacro {
	if uid == "" {
		return nil
	}

	for i := range macros {
		if macros[i].UID == uid {
			return &macros[i]
		}
	}

	return nil
}

// emitParameterItem shows the single "run this macro with this parameter"
// result. autocomplete controls what Tab puts back into the input.
func emitParameterItem(macro *KmMacro, parameter string, autocomplete string) {
	title := parameter
	if title == "" {
		title = macro.Name
	}

	item := wf.NewItem(title).
		UID(macro.UID+"-parameter").
		Arg(macro.UID, parameter).
		Autocomplete(autocomplete).
		Icon(&aw.Icon{Value: "dot.png"})

	if parameter == "" {
		item.Subtitle("Type a parameter, then press Enter").Valid(false)
	} else {
		item.Subtitle(fmt.Sprintf("Run \"%s\"", macro.Name)).Valid(true)
	}

	wf.SendFeedback()
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

	// The macro's identity travels in the query text ("Macro Name -> param"),
	// because that is what Alfred hands back after Tab. A session variable
	// takes over only when the query carries no separator.
	if name, parameter, hasParameter := splitParameterQuery(searchQuery); hasParameter {
		macro := findMacro(macros, name)
		if macro == nil {
			wf.WarnEmpty("No macros found", "Try a different query.")
			wf.SendFeedback()
			return
		}

		emitParameterItem(macro, parameter, name+parameterSeparator+" "+parameter)
		return
	}

	if macro := lookupMacroByUID(macros, os.Getenv(sessionMacroUID)); macro != nil {
		parameter := strings.TrimSpace(searchQuery)
		emitParameterItem(macro, parameter, parameter)
		return
	}

	filteredMacros := SearchMacros(macros, searchQuery)

	var item *aw.Item
	for _, macro := range filteredMacros {
		item = wf.NewItem(macro.Name).
			UID(macro.UID).
			Valid(true).
			Arg(macro.UID).
			Autocomplete(macro.Name+parameterSeparator+" ").
			Var(sessionMacroUID, macro.UID).
			Var(sessionMacroName, macro.Name).
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
