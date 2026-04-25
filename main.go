package main

import (
	"fmt"
	"time"

	"github.com/deanishe/awgo"
)

var wf *aw.Workflow

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

	var item *aw.Item
	for _, macro := range macros {
		item = wf.NewItem(macro.Name).UID(macro.UID).Valid(true).Arg(macro.UID).Icon(&aw.Icon{Value: "dot.png"})
		item.NewModifier("cmd").Subtitle("Execute with parameter...").Arg(macro.UID)
		
		// Ctrl: Copy shell command
		shellCmd := fmt.Sprintf("osascript -e 'tell application \"Keyboard Maestro Engine\" to do script \"%s\"'", macro.UID)
		item.NewModifier("ctrl").Subtitle("Copy shell script trigger").Arg(shellCmd)

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

	args := wf.Args()
	var searchQuery string
	if len(args) > 0 {
		searchQuery = args[0]
	}

	if searchQuery == "" {
		wf.WarnEmpty("No macros found", "It seems that you haven't created any macros yet.")
	} else {
		wf.Filter(searchQuery)
		wf.WarnEmpty("No macros found", "Try a different query.")
	}

	wf.SendFeedback()
}

func init() {
	wf = aw.New()
}

func main() {
	wf.Run(run)
}
