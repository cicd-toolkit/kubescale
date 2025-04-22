/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseScalerAnnotation(val string) (startDay, endDay int, startTime, endTime time.Time, loc *time.Location, err error) {
	// Example: "Mon-Fri 08:00-20:00 UTC"
	parts := strings.Fields(val)
	if len(parts) != 3 {
		err = fmt.Errorf("invalid format, expected: 'Mon-Fri HH:MM-HH:MM TZ'")
		return
	}

	// Parse weekday range
	dayParts := strings.Split(parts[0], "-")
	if len(dayParts) != 2 {
		err = fmt.Errorf("invalid weekday range")
		return
	}
	startDay, err = parseWeekday(dayParts[0])
	if err != nil {
		return
	}
	endDay, err = parseWeekday(dayParts[1])
	if err != nil {
		return
	}

	// Parse time range
	timeParts := strings.Split(parts[1], "-")
	if len(timeParts) != 2 {
		err = fmt.Errorf("invalid time range")
		return
	}
	startTime, err = time.Parse("15:04", timeParts[0])
	if err != nil {
		return
	}
	endTime, err = time.Parse("15:04", timeParts[1])
	if err != nil {
		return
	}

	// Parse timezone
	loc, err = time.LoadLocation(parts[2])
	return
}

func parseWeekday(day string) (int, error) {
	switch strings.ToLower(day) {
	case "sun":
		return 0, nil
	case "mon":
		return 1, nil
	case "tue":
		return 2, nil
	case "wed":
		return 3, nil
	case "thu":
		return 4, nil
	case "fri":
		return 5, nil
	case "sat":
		return 6, nil
	}
	return -1, fmt.Errorf("invalid weekday: %s", day)
}

func isWithinRange(start, end, now time.Time) bool {
	if start.Before(end) {
		return now.After(start) && now.Before(end)
	}
	// Overnight: e.g., 22:00 to 06:00
	return now.After(start) || now.Before(end)
}

func isNowInUptime(startDay, endDay int, start, end time.Time, loc *time.Location) bool {
	now := time.Now().In(loc)
	weekday := int(now.Weekday())
	hourMin := now.Format("15:04")

	// Check day range
	withinDay := startDay <= endDay && weekday >= startDay && weekday <= endDay ||
		startDay > endDay && (weekday >= startDay || weekday <= endDay)

	// Check time range
	nowTime, _ := time.Parse("15:04", hourMin)
	startOnly, _ := time.Parse("15:04", start.Format("15:04"))
	endOnly, _ := time.Parse("15:04", end.Format("15:04"))

	var withinTime bool
	if startOnly.Before(endOnly) {
		withinTime = nowTime.After(startOnly) && nowTime.Before(endOnly)
	} else {
		withinTime = nowTime.After(startOnly) || nowTime.Before(endOnly)
	}

	return withinDay && withinTime
}

func parseHumanDuration(input string) (time.Duration, error) {
	if len(input) < 2 {
		return 0, fmt.Errorf("too short")
	}

	suffix := input[len(input)-1:]
	num := input[:len(input)-1]

	value, err := strconv.Atoi(num)
	if err != nil {
		return 0, err
	}

	switch suffix {
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "M":
		return time.Duration(value) * 30 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown duration suffix: %s", suffix)
	}
}
