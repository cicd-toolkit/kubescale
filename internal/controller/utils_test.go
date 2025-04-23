package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var timeNow = time.Now

func TestParseHumanDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"10m", 10 * time.Minute, false},
		{"2h", 2 * time.Hour, false},
		{"1d", 24 * time.Hour, false},
		{"5x", 0, true}, // Invalid suffix
		{"", 0, true},   // Too short
		{"m", 0, true},  // Missing number
	}

	for _, test := range tests {
		result, err := parseHumanDuration(test.input)
		if test.hasError {
			assert.Error(t, err, "expected an error for input: %s", test.input)
		} else {
			assert.NoError(t, err, "did not expect an error for input: %s", test.input)
			assert.Equal(t, test.expected, result, "unexpected result for input: %s", test.input)
		}
	}
}
func TestParseScalerAnnotation(t *testing.T) {
	tests := []struct {
		input       string
		startDay    int
		endDay      int
		startTime   string
		endTime     string
		location    string
		expectError bool
	}{
		{"Mon-Fri 08:00-20:00 UTC", 1, 5, "08:00", "20:00", "UTC", false},
		{"Sat-Sun 22:00-06:00 UTC", 6, 0, "22:00", "06:00", "UTC", false},
		{"Mon-Fri 08:00-20:00 InvalidTZ", 0, 0, "", "", "", true},
		{"InvalidFormat", 0, 0, "", "", "", true},
		{"Mon-Fri InvalidTimeRange UTC", 0, 0, "", "", "", true},
		{"InvalidDayRange 08:00-20:00 UTC", 0, 6, "08:00", "20:00", "UTC", false},
		{"08:00-20:00 UTC", 0, 6, "08:00", "20:00", "UTC", false},     // Default days
		{"Mon-Fri 08:00-20:00", 1, 5, "08:00", "20:00", "UTC", false}, // Default timezone
	}

	for _, test := range tests {
		timerange, err := parseScalerAnnotation(test.input)
		if test.expectError {
			assert.Error(t, err, "expected an error for input: %s", test.input)
		} else {
			assert.NoError(t, err, "did not expect an error for input: %s", test.input)
			assert.Equal(t, test.startDay, int(timerange.StartDay), "unexpected start day for input: %s", test.input)
			assert.Equal(t, test.endDay, int(timerange.EndDay), "unexpected end day for input: %s", test.input)
			assert.Equal(t, test.startTime, timerange.Start.Format("15:04"), "unexpected start time for input: %s", test.input)
			assert.Equal(t, test.endTime, timerange.End.Format("15:04"), "unexpected end time for input: %s", test.input)
			assert.Equal(t, test.location, timerange.Location.String(), "unexpected location for input: %s", test.input)
		}
	}
}

func TestParseWeekday(t *testing.T) {
	tests := []struct {
		input       string
		expected    int
		expectError bool
	}{
		{"Mon", 1, false},
		{"Fri", 5, false},
		{"Sun", 0, false},
		{"InvalidDay", 0, false}, // Invalid day, should return  0
	}

	for _, test := range tests {
		result := parseWeekday(test.input)
		if test.expectError {
			assert.Equal(t, -1, result, "expected an error for input: %s", test.input)
		} else {
			assert.Equal(t, time.Weekday(test.expected), result, "unexpected result for input: %s", test.input)
		}
	}
}
func TestIsInRange(t *testing.T) {
	tests := []struct {
		startDay      time.Weekday
		endDay        time.Weekday
		startTime     string
		endTime       string
		location      string
		currentTime   string
		expectInRange bool
	}{
		{time.Monday, time.Friday, "08:00", "18:00", "UTC", "2023-10-03T10:00:00Z", true},        // Within range
		{time.Monday, time.Friday, "08:00", "18:00", "UTC", "2023-10-02T19:00:00Z", false},       // Outside time range
		{time.Monday, time.Friday, "18:00", "08:00", "UTC", "2023-10-02T19:00:00Z", true},        // Overnight range
		{time.Monday, time.Friday, "08:00", "18:00", "UTC", "2023-10-07T09:00:00Z", false},       // Outside day range
		{time.Monday, time.Friday, "08:00", "18:00", "InvalidTZ", "2023-10-02T09:00:00Z", false}, // Invalid timezone
	}

	for _, test := range tests {
		loc, err := time.LoadLocation(test.location)
		if test.location == "InvalidTZ" {
			assert.Error(t, err, "expected an error for invalid timezone")
			continue
		}
		assert.NoError(t, err, "did not expect an error for valid timezone")

		start, _ := time.Parse("15:04", test.startTime)
		end, _ := time.Parse("15:04", test.endTime)
		current, _ := time.Parse(time.RFC3339, test.currentTime)

		tr := &TimeRange{
			StartDay: test.startDay,
			EndDay:   test.endDay,
			Start:    start,
			End:      end,
			Location: loc,
		}
		// Mock the Now function to return the current time
		result := tr.isInRange(current)
		assert.Equal(t, test.expectInRange, result, "unexpected result for test case: %+v", test)
	}
}
