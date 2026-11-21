package main

import "testing"

func TestStripMacOSPrefix(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", "Unknown"},
		{"macOS 15.1", "15.1"},
		{"macOS 14.7.2 (23H311)", "14.7.2 (23H311)"},
		{"15.1", "15.1"},
		{"macOS", "macOS"},
		{"MacOS 15.1", "MacOS 15.1"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := stripMacOSPrefix(tt.in); got != tt.want {
				t.Errorf("stripMacOSPrefix(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
