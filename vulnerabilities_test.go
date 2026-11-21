package main

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
)

func TestVulnerabilitySeverity(t *testing.T) {
	tests := []struct {
		name  string
		score *float64
		want  severity
	}{
		{"nil", nil, "Unscored"},
		{"below range", new(-0.1), "Unscored"},
		{"far below range", new(-50.0), "Unscored"},
		{"above range", new(10.1), "Unscored"},
		{"infinity", new(math.Inf(1)), "Unscored"},
		{"0", new(0.0), "None"},
		{"0.1", new(0.1), "Low"},
		{"3.9", new(3.9), "Low"},
		{"4.0", new(4.0), "Medium"},
		{"6.9", new(6.9), "Medium"},
		{"7.0", new(7.0), "High"},
		{"8.9", new(8.9), "High"},
		{"9.0", new(9.0), "Critical"},
		{"10.0", new(10.0), "Critical"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vulnerabilitySeverity(tt.score); got != tt.want {
				t.Errorf("vulnerabilitySeverity(%s) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestFilterVulnerabilities(t *testing.T) {
	openssl := workbrew.VulnerabilityReport{
		Package:         "openssl",
		AffectedDevices: []string{"abc123", "XYZ999"},
		Vulnerabilities: []workbrew.CVE{
			{CleanID: "CVE-CRIT", CVSSScore: new(9.8)},
			{CleanID: "CVE-MED", CVSSScore: new(5.0)},
		},
	}
	curl := workbrew.VulnerabilityReport{
		Package:         "curl",
		AffectedDevices: []string{"XYZ999"},
		Vulnerabilities: []workbrew.CVE{
			{CleanID: "CVE-HIGH", CVSSScore: new(7.5)},
			{CleanID: "CVE-NONE", CVSSScore: nil},
		},
	}
	all := []workbrew.VulnerabilityReport{openssl, curl}

	type result struct {
		Package string
		Devices []string
		CVEs    []string
	}
	summarise := func(items []workbrew.VulnerabilityReport) []result {
		out := []result{}
		for _, item := range items {
			r := result{Package: item.Package, Devices: item.AffectedDevices}
			for _, cve := range item.Vulnerabilities {
				r.CVEs = append(r.CVEs, cve.CleanID)
			}
			out = append(out, r)
		}
		return out
	}

	tests := []struct {
		name     string
		severity severity
		device   string
		want     []result
	}{
		{
			name: "no filters",
			want: []result{
				{"openssl", []string{"abc123", "XYZ999"}, []string{"CVE-CRIT", "CVE-MED"}},
				{"curl", []string{"XYZ999"}, []string{"CVE-HIGH", "CVE-NONE"}},
			},
		},
		{
			name:     "severity only",
			severity: severityHigh,
			want:     []result{{"curl", []string{"XYZ999"}, []string{"CVE-HIGH"}}},
		},
		{
			name:     "severity with no matches",
			severity: severityLow,
			want:     []result{},
		},
		{
			name:   "device exact, case-insensitive, narrows affected devices",
			device: "ABC123",
			want:   []result{{"openssl", []string{"abc123"}, []string{"CVE-CRIT", "CVE-MED"}}},
		},
		{
			name:   "device partial does not match",
			device: "abc",
			want:   []result{},
		},
		{
			name:     "severity and device",
			severity: severityCritical,
			device:   "xyz999",
			want:     []result{{"openssl", []string{"XYZ999"}, []string{"CVE-CRIT"}}},
		},
		{
			name:     "severity and device with no overlap",
			severity: severityHigh,
			device:   "abc123",
			want:     []result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterVulnerabilities(all, tt.severity, tt.device)
			if !reflect.DeepEqual(summarise(got), tt.want) {
				t.Errorf("filterVulnerabilities(%q, %q) =\n  %+v\nwant\n  %+v", tt.severity, tt.device, summarise(got), tt.want)
			}

			encoded, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			if len(tt.want) == 0 && string(encoded) != "[]" {
				t.Errorf("JSON = %s, want []", encoded)
			}
		})
	}

	// Filtering must not modify the input.
	if len(openssl.AffectedDevices) != 2 || len(openssl.Vulnerabilities) != 2 {
		t.Errorf("input was modified: %+v", openssl)
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		in      string
		want    severity
		wantErr bool
	}{
		{"", "", false},
		{"critical", severityCritical, false},
		{"HIGH", severityHigh, false},
		{" Medium ", severityMedium, false},
		{"low", severityLow, false},
		{"none", "", true},
		{"unscored", "", true},
		{"urgent", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := parseSeverity(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseSeverity(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseSeverity(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
