package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/mrbrown89/workbrew-cli/internal/workbrew"
)

func TestFilterAnalytics(t *testing.T) {
	items := []workbrew.AnalyticsItem{
		{Device: "ABC123", Command: "a"},
		{Device: "ABC1234", Command: "b"},
		{Device: "abc123", Command: "c"},
	}

	tests := []struct {
		device string
		want   []string
	}{
		{"", []string{"a", "b", "c"}},
		{"ABC123", []string{"a", "c"}},
		{"abc", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.device, func(t *testing.T) {
			got := []string{}
			filtered := filterAnalytics(items, tt.device)
			for _, item := range filtered {
				got = append(got, item.Command)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("filterAnalytics(%q) = %v, want %v", tt.device, got, tt.want)
			}
			if encoded, _ := json.Marshal(filtered); len(tt.want) == 0 && string(encoded) != "[]" {
				t.Errorf("JSON = %s, want []", encoded)
			}
		})
	}
}

func TestSortAnalytics(t *testing.T) {
	items := []workbrew.AnalyticsItem{
		{Device: "B", Command: "brew upgrade", Count: 3},
		{Device: "A", Command: "brew upgrade", Count: 3},
		{Device: "Z", Command: "brew install", Count: 3},
		{Device: "A", Command: "brew doctor", Count: 9},
		{Device: "A", Command: "brew cleanup", Count: 1},
	}

	sortAnalytics(items)

	var got []string
	for _, item := range items {
		got = append(got, fmt.Sprintf("%d %s %s", item.Count, item.Command, item.Device))
	}
	want := []string{
		"9 brew doctor A",
		"3 brew install Z",
		"3 brew upgrade A",
		"3 brew upgrade B",
		"1 brew cleanup A",
	}
	if !slices.Equal(got, want) {
		t.Errorf("sortAnalytics =\n  %v\nwant\n  %v", got, want)
	}
}

func TestSortBrewCommands(t *testing.T) {
	items := []workbrew.BrewCommand{
		{Command: "brew update", RunCount: 2},
		{Command: "brew cleanup", RunCount: 2},
		{Command: "brew upgrade", RunCount: 5},
	}

	sortBrewCommands(items)

	var got []string
	for _, item := range items {
		got = append(got, item.Command)
	}
	if want := []string{"brew upgrade", "brew cleanup", "brew update"}; !slices.Equal(got, want) {
		t.Errorf("sortBrewCommands = %v, want %v", got, want)
	}
}
