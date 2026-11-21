package main

import (
	"fmt"
	"time"
)

// clockSkewTolerance is how far in the future a timestamp may be and still
// count as "Today", to allow for clocks that are slightly ahead.
const clockSkewTolerance = 5 * time.Minute

// daysAgo describes how long before now dateString (RFC 3339) was:
// "Today", "1 day" or "N days". Empty and unparsable dates, and dates more
// than clockSkewTolerance in the future, are "Unknown".
func daysAgo(dateString string, now time.Time) string {
	if dateString == "" || dateString == "Never" {
		return "Unknown"
	}

	parsedDate, err := time.Parse(time.RFC3339, dateString)
	if err != nil || parsedDate.After(now.Add(clockSkewTolerance)) {
		return "Unknown"
	}
	if parsedDate.After(now) {
		return "Today"
	}

	switch days := int(now.Sub(parsedDate).Hours() / 24); days {
	case 0:
		return "Today"
	case 1:
		return "1 day"
	default:
		return fmt.Sprintf("%d days", days)
	}
}
