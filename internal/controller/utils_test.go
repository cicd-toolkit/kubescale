package controller

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
		{"InvalidDayRange 08:00-20:00 UTC", 0, 0, "", "", "", true},
	}

	for _, test := range tests {
		startDay, endDay, startTime, endTime, loc, err := parseScalerAnnotation(test.input)
		if test.expectError {
			assert.Error(t, err, "expected an error for input: %s with ", test.input)
		} else {
			assert.NoError(t, err, "did not expect an error for input: %s", test.input)
			assert.Equal(t, test.startDay, startDay, "unexpected startDay for input: %s", test.input)
			assert.Equal(t, test.endDay, endDay, "unexpected endDay for input: %s", test.input)
			assert.Equal(t, test.startTime, startTime.Format("15:04"), "unexpected startTime for input: %s", test.input)
			assert.Equal(t, test.endTime, endTime.Format("15:04"), "unexpected endTime for input: %s", test.input)
			assert.Equal(t, test.location, loc.String(), "unexpected location for input: %s", test.input)
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
		{"InvalidDay", -1, true},
	}

	for _, test := range tests {
		result, err := parseWeekday(test.input)
		if test.expectError {
			assert.Error(t, err, "expected an error for input: %s", test.input)
		} else {
			assert.NoError(t, err, "did not expect an error for input: %s", test.input)
			assert.Equal(t, test.expected, result, "unexpected result for input: %s", test.input)
		}
	}
}

func TestIsWithinRange(t *testing.T) {
	tests := []struct {
		start    string
		end      string
		now      string
		expected bool
	}{
		{"08:00", "20:00", "12:00", true},
		{"08:00", "20:00", "07:00", false},
		{"22:00", "06:00", "23:00", true},
		{"22:00", "06:00", "05:00", true},
		{"22:00", "06:00", "07:00", false},
	}

	for _, test := range tests {
		start, _ := time.Parse("15:04", test.start)
		end, _ := time.Parse("15:04", test.end)
		now, _ := time.Parse("15:04", test.now)

		result := isWithinRange(start, end, now)
		assert.Equal(t, test.expected, result, "unexpected result for start: %s, end: %s, now: %s", test.start, test.end, test.now)
	}
}
