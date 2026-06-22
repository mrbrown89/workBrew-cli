package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const workbrewAPIVersion = "v0"

type Device struct {
	SerialNumber string `json:"serial_number"`
	AssignedUser string `json:"mdm_user_or_device_name"`
	DeviceType   string `json:"device_type"`
	OSVersion    string `json:"os_version"`
	LastSeenAt   string `json:"last_seen_at"`
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

type CVE struct {
	CleanID   string  `json:"clean_id"`
	CVSSScore float64 `json:"cvss_score"`
}

type VulnerabilityReport struct {
	Vulnerabilities     []CVE    `json:"vulnerabilities"`
	Formula             string   `json:"formula"`
	OutdatedDevices     []string `json:"outdated_devices"`
	Supported           bool     `json:"supported"`
	HomebrewCoreVersion string   `json:"homebrew_core_version"`
}

type AnalyticsItem struct {
	Device  string `json:"device"`
	Command string `json:"command"`
	LastRun string `json:"last_run"`
	Count   int    `json:"count"`
}

type Brewfile struct {
	Label             string   `json:"label"`
	Slug              string   `json:"slug"`
	Content           string   `json:"content"`
	LastUpdatedByUser string   `json:"last_updated_by_user"`
	StartedAt         string   `json:"started_at"`
	FinishedAt        string   `json:"finished_at"`
	Devices           []string `json:"devices"`
	RunCount          int      `json:"run_count"`
}

type BrewCommand struct {
	Command           string   `json:"command"`
	Label             string   `json:"label"`
	LastUpdatedByUser string   `json:"last_updated_by_user"`
	StartedAt         string   `json:"started_at"`
	FinishedAt        string   `json:"finished_at"`
	Devices           []string `json:"devices"`
	RunCount          int      `json:"run_count"`
}

type BrewCommandCreateRequest struct {
	Arguments string `json:"arguments"`
}

type BrewCommandCreateResponse struct {
	Message string `json:"message"`
}

func workbrewGetJSON(config Config, token string, endpoint string, target any) error {
	url := fmt.Sprintf("%s/%s", config.URL, endpoint)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Workbrew API returned %s", response.Status)
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func getDevices(config Config, token string) ([]Device, error) {
	var devices []Device

	if err := workbrewGetJSON(config, token, "devices.json", &devices); err != nil {
		return nil, err
	}

	return devices, nil
}

func getFormulae(config Config, token string) ([]Formula, error) {
	var formulae []Formula

	if err := workbrewGetJSON(config, token, "formulae.json", &formulae); err != nil {
		return nil, err
	}

	return formulae, nil
}

func getCasks(config Config, token string) ([]Cask, error) {
	var casks []Cask

	if err := workbrewGetJSON(config, token, "casks.json", &casks); err != nil {
		return nil, err
	}

	return casks, nil
}

func getVulnerabilities(config Config, token string) ([]VulnerabilityReport, error) {
	var vulnerabilities []VulnerabilityReport

	if err := workbrewGetJSON(config, token, "vulnerabilities.json", &vulnerabilities); err != nil {
		return nil, err
	}

	return vulnerabilities, nil
}

func getAnalytics(config Config, token string) ([]AnalyticsItem, error) {
	var analytics []AnalyticsItem

	if err := workbrewGetJSON(config, token, "analytics.json", &analytics); err != nil {
		return nil, err
	}

	return analytics, nil
}

func getBrewfiles(config Config, token string) ([]Brewfile, error) {
	var brewfiles []Brewfile

	if err := workbrewGetJSON(config, token, "brewfiles.json", &brewfiles); err != nil {
		return nil, err
	}

	return brewfiles, nil
}

func getBrewCommands(config Config, token string) ([]BrewCommand, error) {
	var brewCommands []BrewCommand

	if err := workbrewGetJSON(config, token, "brew_commands.json", &brewCommands); err != nil {
		return nil, err
	}

	return brewCommands, nil
}

func workbrewPostJSON(config Config, token string, endpoint string, payload any, target any) error {

	url := fmt.Sprintf("%s/%s", config.URL, endpoint)
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-Workbrew-API-Version", workbrewAPIVersion)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Workbrew API returned %s", response.Status)
	}
	if target != nil {
		if err := json.NewDecoder(response.Body).Decode(target); err != nil {
			return err
		}
	}
	return nil

}

func createBrewCommand(config Config, token string, arguments string) (BrewCommandCreateResponse, error) {

	var result BrewCommandCreateResponse
	request := BrewCommandCreateRequest{
		Arguments: arguments,
	}
	if err := workbrewPostJSON(config, token, "brew_commands.json", request, &result); err != nil {
		return result, err
	}
	return result, nil

}
