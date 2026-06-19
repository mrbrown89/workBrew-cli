package main

func countFormulaeForDevice(serial string, formulae []Formula) int {
	count := 0

	for _, formula := range formulae {
		for _, device := range formula.Devices {
			if device == serial {
				count++
				break
			}
		}
	}

	return count
}

func countCasksForDevice(serial string, casks []Cask) int {
	count := 0

	for _, cask := range casks {
		for _, device := range cask.Devices {
			if device == serial {
				count++
				break
			}
		}
	}

	return count
}

func countOutdatedForDevice(
	serial string,
	formulae []Formula,
	casks []Cask,
) int {
	count := 0

	for _, formula := range formulae {
		if !formula.Outdated {
			continue
		}

		for _, device := range formula.Devices {
			if device == serial {
				count++
				break
			}
		}
	}

	for _, cask := range casks {
		if !cask.Outdated {
			continue
		}

		for _, device := range cask.Devices {
			if device == serial {
				count++
				break
			}
		}
	}

	return count
}
