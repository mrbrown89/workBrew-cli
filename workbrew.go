package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const workbrewAPIVersion = "v0"

func workbrewGet(config Config, token string, endpoint string) (any, error) {
	url := fmt.Sprintf("%s/%s", config.URL, endpoint)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Workbrew API returned %s", response.Status)
	}

	var result any

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

type Device struct {
	SerialNumber string `json:"serial_number"`
	AssignedUser string `json:"mdm_user_or_device_name"`
	DeviceType   string `json:"device_type"`
	OSVersion    string `json:"os_version"`
	LastSeenAt   string `json:"last_seen_at"`
}

func getDevices(config Config, token string) ([]Device, error) {
	url := fmt.Sprintf("%s/%s", config.URL, "devices.json")

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Workbrew API returned %s", response.Status)
	}

	var devices []Device

	if err := json.NewDecoder(response.Body).Decode(&devices); err != nil {
		return nil, err
	}

	return devices, nil
}

type Formula struct {
	Name                string   `json:"name"`
	HomebrewCoreVersion string   `json:"homebrew_core_version"`
	Outdated            bool     `json:"outdated"`
	Devices             []string `json:"devices"`
}

type Cask struct {
	Name                string   `json:"name"`
	HomebrewCaskVersion string   `json:"homebrew_cask_version"`
	Outdated            bool     `json:"outdated"`
	Devices             []string `json:"devices"`
}

func getFormulae(config Config, token string) ([]Formula, error) {
	url := fmt.Sprintf("%s/%s", config.URL, "formulae.json")

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Workbrew API returned %s", response.Status)
	}

	var formulae []Formula

	if err := json.NewDecoder(response.Body).Decode(&formulae); err != nil {
		return nil, err
	}

	return formulae, nil
}

func getCasks(config Config, token string) ([]Cask, error) {
	url := fmt.Sprintf("%s/%s", config.URL, "casks.json")

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Workbrew API returned %s", response.Status)
	}

	var casks []Cask

	if err := json.NewDecoder(response.Body).Decode(&casks); err != nil {
		return nil, err
	}

	return casks, nil
}
