package workbrew

// API endpoints, relative to the workspace API base URL.
const (
	endpointDevices         = "devices.json"
	endpointFormulae        = "formulae.json"
	endpointCasks           = "casks.json"
	endpointVulnerabilities = "vulnerabilities.json"
	endpointAnalytics       = "analytics.json"
	endpointBrewfiles       = "brewfiles.json"
	endpointBrewCommands    = "brew_commands.json"
)

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
	CleanID   string   `json:"clean_id"`
	CVSSScore *float64 `json:"cvss_score"`
}

type VulnerabilityReport struct {
	Vulnerabilities []CVE    `json:"vulnerabilities"`
	Package         string   `json:"package"`
	DisplayName     string   `json:"display_name"`
	PackageType     string   `json:"package_type"`
	AffectedDevices []string `json:"affected_devices"`
	LatestVersion   *string  `json:"latest_version"`
	Supported       *bool    `json:"supported"`
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
