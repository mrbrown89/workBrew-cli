package main

func getFormulaeForDevice(
	serial string,
	formulae []Formula,
) []Formula {

	var results []Formula

	for _, formula := range formulae {
		for _, device := range formula.Devices {
			if device == serial {
				results = append(results, formula)
				break
			}
		}
	}

	return results
}

func getCasksForDevice(
	serial string,
	casks []Cask,
) []Cask {

	var results []Cask

	for _, cask := range casks {
		for _, device := range cask.Devices {
			if device == serial {
				results = append(results, cask)
				break
			}
		}
	}

	return results
}
