package main

import (
	"encoding/xml"
	"errors"
	"os"
	"os/exec"
)

type KmItem struct {
	Keys   []string `xml:"key"`
	Values []string `xml:",any"`
}

type KmCategory struct {
	Keys   []string `xml:"key"`
	Values []string `xml:"string"`
	Items  []KmItem `xml:"array>dict"`
}

type KmCategories struct {
	Categories []KmCategory `xml:"array>dict"`
}

type KmMacro struct {
	UID      string
	Name     string
	Category string
	TriggerString string
	Hotkey   string
}

func getKmMacros() (map[string]KmMacro, error) {
	// Execute the command to get KM macros (usually an AppleScript)
	cmd := "osascript ./get_hotkey_km_macros.scpt"

	// Allow override through environment variable for testing
	if envCmd := os.Getenv("GET_ALL_KM_MACROS_COMMAND"); envCmd != "" {
		cmd = envCmd
	}

	categories, err := getKmCategories(cmd)
	if err != nil {
		return nil, err
	}

	macros := make(map[string]KmMacro)
	var uid string

	// Process all macros from the XML output
	for _, category := range categories.Categories {
		for _, item := range category.Items {
			uid = item.getValueByKey("uid")
			macros[uid] = KmMacro{
				UID:           uid,
				Name:          item.getValueByKey("name"),
				Category:      category.getValueByKey("name"),
				Hotkey:        item.getValueByKey("key"), 
				TriggerString: item.getValueByKey("triggerstring"),
			}
		}
	}

	return macros, nil
}

func getKmCategories(cmdStr string) (KmCategories, error) {
	// Execute command using sh
	out, err := exec.Command("sh", "-c", cmdStr).Output()
	if err != nil {
		return KmCategories{}, errors.New("Unable to execute Keyboard Maestro command: " + err.Error())
	}

	var categories KmCategories
	err = xml.Unmarshal(out, &categories)
	if err != nil {
		return categories, errors.New("Unable to parse Keyboard Maestro XML data")
	}

	return categories, nil
}

func (item KmItem) getValueByKey(requestedKey string) string {
	for i, key := range item.Keys {
		if key == requestedKey {
			return item.Values[i]
		}
	}

	return ""
}

// TODO Find out how to use the same func for both KmItem and KmCategory
func (item KmCategory) getValueByKey(requestedKey string) string {
	for i, key := range item.Keys {
		if key == requestedKey {
			return item.Values[i]
		}
	}

	return ""
}
