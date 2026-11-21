package main

import (
	"testing"
	"time"
)

func TestDaysAgo(t *testing.T) {
	now := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	at := func(d time.Duration) string { return now.Add(d).Format(time.RFC3339) }

	tests := []struct {
		name string
		date string
		want string
	}{
		{"empty", "", "Unknown"},
		{"never", "Never", "Unknown"},
		{"garbage", "yesterday", "Unknown"},
		{"date only", "2026-06-14", "Unknown"},
		{"now", at(0), "Today"},
		{"23h59m ago", at(-23*time.Hour - 59*time.Minute), "Today"},
		{"exactly 24h ago", at(-24 * time.Hour), "1 day"},
		{"47h ago", at(-47 * time.Hour), "1 day"},
		{"48h ago", at(-48 * time.Hour), "2 days"},
		{"30 days ago", at(-30 * 24 * time.Hour), "30 days"},
		{"other time zone", "2026-06-14T13:00:00+01:00", "1 day"},
		{"1s in the future", at(time.Second), "Today"},
		{"exactly 5m in the future", at(5 * time.Minute), "Today"},
		{"5m1s in the future", at(5*time.Minute + time.Second), "Unknown"},
		{"a day in the future", at(24 * time.Hour), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := daysAgo(tt.date, now); got != tt.want {
				t.Errorf("daysAgo(%q) = %q, want %q", tt.date, got, tt.want)
			}
		})
	}
}
