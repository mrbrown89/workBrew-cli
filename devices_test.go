package main

import (
	"slices"
	"testing"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
)

func TestMatchDevices(t *testing.T) {
	devices := []workbrew.Device{
		{SerialNumber: "ABC123", AssignedUser: "Alex Example"},
		{SerialNumber: "ABC1234", AssignedUser: "Alexandra Person"},
		{SerialNumber: "XYZ999", AssignedUser: "Sam"},
		{SerialNumber: "QRS111", AssignedUser: "Shared"},
		{SerialNumber: "QRS222", AssignedUser: "shared"},
	}

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"exact serial beats partial", "ABC123", []string{"ABC123"}},
		{"exact serial is case-insensitive", "abc123", []string{"ABC123"}},
		{"exact user", "sam", []string{"XYZ999"}},
		{"single partial serial", "xyz", []string{"XYZ999"}},
		{"single partial user", "example", []string{"ABC123"}},
		{"several partial", "abc12", []string{"ABC123", "ABC1234"}},
		{"several partial users", "alex", []string{"ABC123", "ABC1234"}},
		{"several exact users", "SHARED", []string{"QRS111", "QRS222"}},
		{"none", "nobody", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, device := range matchDevices(devices, tt.query) {
				got = append(got, device.SerialNumber)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("matchDevices(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestBuildDeviceDetails(t *testing.T) {
	device := workbrew.Device{SerialNumber: "ABC123"}
	formulae := []workbrew.Formula{
		{Name: "zsh", Devices: []string{"ABC123"}},
		{Name: "wget", Outdated: true, Devices: []string{"ABC123"}},
		{Name: "jq", Outdated: true, Devices: []string{"OTHER"}},
		{Name: "git", Devices: []string{"OTHER", "ABC123"}},
	}
	casks := []workbrew.Cask{{Name: "firefox", Devices: []string{"OTHER"}}}

	details := buildDeviceDetails(device, formulae, casks, true)

	if details.FormulaCount != 3 || details.CaskCount != 0 || details.OutdatedCount != 1 {
		t.Errorf("counts = %d/%d/%d, want 3/0/1", details.FormulaCount, details.CaskCount, details.OutdatedCount)
	}

	var names []string
	for _, formula := range details.Formulae {
		names = append(names, formula.Name)
	}
	if want := []string{"wget", "git", "zsh"}; !slices.Equal(names, want) {
		t.Errorf("formulae order = %v, want %v (outdated first, then by name)", names, want)
	}
	if details.Casks == nil {
		t.Error("casks = nil with --apps, want empty slice")
	}

	without := buildDeviceDetails(device, formulae, casks, false)
	if without.Formulae != nil || without.Casks != nil {
		t.Error("formulae/casks set without --apps")
	}
}
