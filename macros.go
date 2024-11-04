package main

import (
	"encoding/xml"
	"errors"
	"os"
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
	Hotkey   string
}

func getKmMacros() (map[string]KmMacro, error) {
	// Replace AppleScript execution with direct file reading
	xmlPath := "/Applications/Alfred 5.app/Contents/Resources/km_data.xml"

	// Allow override through environment variable for testing
	if envPath := os.Getenv("KM_XML_PATH"); envPath != "" {
		xmlPath = envPath
	}

	categories, err := getKmCategories(xmlPath)
	if err != nil {
		return nil, err
	}

	macros := make(map[string]KmMacro)
	var uid string

	// Process all macros from the single XML file
	for _, category := range categories.Categories {
		for _, item := range category.Items {
			uid = item.getValueByKey("uid")
			macros[uid] = KmMacro{
				UID:      uid,
				Name:     item.getValueByKey("name"),
				Category: category.getValueByKey("name"),
				Hotkey:   item.getValueByKey("key"), // Assuming hotkey is in the same XML
			}
		}
	}

	return macros, nil
}

func getKmCategories(filePath string) (KmCategories, error) {
	// Read XML file instead of executing command
	xmlData, err := os.ReadFile(filePath)
	if err != nil {
		return KmCategories{}, errors.New("Unable to read Keyboard Maestro XML file")
	}

	var categories KmCategories
	err = xml.Unmarshal(xmlData, &categories)
	if err != nil {
		return categories, errors.New("Unable to parse Keyboard Maestro XML file")
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
