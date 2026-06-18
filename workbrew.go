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
