package main

import (
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

type PackageRequest struct {
	ID             string  `json:"id"`
	PackageName    string  `json:"package_name"`
	PackageType    string  `json:"package_type"`
	Tap            string  `json:"tap"`
	Status         string  `json:"status"`
	Device         string  `json:"device"`
	ScopeType      string  `json:"scope_type"`
	ScopeID        *string `json:"scope_id"`
	DecidedBy      string  `json:"decided_by"`
	DecisionReason string  `json:"decision_reason"`
	DecidedAt      *string `json:"decided_at"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
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

func getPackageRequests(config Config, token string) ([]PackageRequest, error) {
	var packageRequests []PackageRequest

	if err := workbrewGetJSON(config, token, "package_requests.json", &packageRequests); err != nil {
		return nil, err
	}

	return packageRequests, nil
}
