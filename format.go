package main

import "strings"

func stripMacOSPrefix(osVersion string) string {
	if osVersion == "" {
		return "Unknown"
	}

	return strings.TrimPrefix(osVersion, "macOS ")
}
