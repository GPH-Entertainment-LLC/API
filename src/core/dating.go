package core

import (
	"fmt"
	"time"
)

func GenerateDaysOfMonth(year int64, month int64) []string {
	// Get the first day of the specified month and year
	firstDay := time.Date(int(year), time.Month(month), 1, 0, 0, 0, 0, time.Local)

	// Calculate the last day of the specified month and year
	lastDay := firstDay.AddDate(0, 1, -1)

	// Generate a list of strings for each day in the month
	var days []string
	for day := firstDay; !day.After(lastDay); day = day.AddDate(0, 0, 1) {
		days = append(days, day.Format("Jan 2"))
	}

	return days
}

func GenerateMonthStrings() []string {
	var months []string
	layout := "Jan"

	// Iterate over 12 months
	for month := time.January; month <= time.December; month++ {
		months = append(months, time.Date(2000, month+1, 0, 0, 0, 0, 0, time.UTC).Format(layout))
	}

	return months
}

func GenerateLast5Years() []string {
	var years []string

	// Get the current year
	currentYear := time.Now().Year()

	// Iterate over the last 5 years
	for i := 4; i >= 0; i-- {
		year := currentYear - i
		years = append(years, fmt.Sprint(year))
	}

	return years
}
