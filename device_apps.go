package main

import (
	"slices"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
)

// deviceApps is what one device has installed. The lists are never nil.
type deviceApps struct {
	formulae []workbrew.Formula
	casks    []workbrew.Cask
	outdated int
}

// appsForDevice collects the formulae and casks installed on serial and
// counts how many of them are outdated, in one pass over each list.
func appsForDevice(serial string, formulae []workbrew.Formula, casks []workbrew.Cask) deviceApps {
	apps := deviceApps{formulae: []workbrew.Formula{}, casks: []workbrew.Cask{}}

	for _, formula := range formulae {
		if !slices.Contains(formula.Devices, serial) {
			continue
		}
		apps.formulae = append(apps.formulae, formula)
		if formula.Outdated {
			apps.outdated++
		}
	}

	for _, cask := range casks {
		if !slices.Contains(cask.Devices, serial) {
			continue
		}
		apps.casks = append(apps.casks, cask)
		if cask.Outdated {
			apps.outdated++
		}
	}

	return apps
}
