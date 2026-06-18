package main

import (
	"fmt"
	"time"
)

func daysAgo(dateString string) string {
	if dateString == "" || dateString == "Never" {
		return "Unknown"
	}

	parsedDate, err := time.Parse(time.RFC3339, dateString)
	if err != nil {
		return "Unknown"
	}

	days := int(time.Since(parsedDate).Hours() / 24)

	if days == 0 {
		return "Today"
	}

	return fmt.Sprintf("%d", days)
}
