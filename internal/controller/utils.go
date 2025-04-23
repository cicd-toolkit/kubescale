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
	"regexp"
	"strconv"
	"strings"
	"time"
)

type TimeRange struct {
	StartDay time.Weekday
	EndDay   time.Weekday
	Start    time.Time
	End      time.Time
	Location *time.Location
}

func parseScalerAnnotation(annot string) (*TimeRange, error) {
	var timeRangeRegex = regexp.MustCompile(`(?:(\w{3})-(\w{3})\s+)?(\d{2}:\d{2})-(\d{2}:\d{2})(?:\s+([\w/_+-]+))?`)

	matches := timeRangeRegex.FindStringSubmatch(annot)
	if matches == nil {
		return nil, fmt.Errorf("invalid format: %s", annot)
	}

	startDayStr := matches[1]
	endDayStr := matches[2]
	startTimeStr := matches[3]
	endTimeStr := matches[4]
	timezone := matches[5]

	// Default days if not specified
	startDay := time.Sunday
	endDay := time.Saturday
	if startDayStr != "" && endDayStr != "" {
		startDay = parseWeekday(startDayStr)
		endDay = parseWeekday(endDayStr)
	}

	// Default timezone if not specified
	loc := time.UTC
	if timezone != "" {
		var err error
		loc, err = time.LoadLocation(timezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone: %s", timezone)
		}
	}

	start, err := parseHourMin(startTimeStr)
	if err != nil {
		return nil, err
	}
	end, err := parseHourMin(endTimeStr)
	if err != nil {
		return nil, err
	}

	return &TimeRange{
		StartDay: startDay,
		EndDay:   endDay,
		Start:    start,
		End:      end,
		Location: loc,
	}, nil
}

func (tr *TimeRange) isInRange(t time.Time) bool {
	now := time.Now().In(tr.Location)
	if !t.IsZero() {
		fmt.Printf("Using provided time: %s\n", t)
		now = t.In(tr.Location)
	}
	weekday := int(now.Weekday())
	fmt.Printf("Current time: %s, Weekday: %d\n", now.Format("15:04"), weekday)

	// Check day range
	withinDay := int(tr.StartDay) <= int(tr.EndDay) && weekday >= int(tr.StartDay) && weekday <= int(tr.EndDay) ||
		int(tr.StartDay) > int(tr.EndDay) && (weekday >= int(tr.StartDay) || weekday <= int(tr.EndDay))

	// Check time range
	nowTime, _ := time.Parse("15:04", now.Format("15:04"))
	var withinTime bool
	if tr.Start.Before(tr.End) {
		withinTime = nowTime.After(tr.Start) && nowTime.Before(tr.End)
	} else {
		withinTime = nowTime.After(tr.Start) || nowTime.Before(tr.End)
	}

	return withinDay && withinTime
}

func parseHourMin(s string) (time.Time, error) {
	t, err := time.Parse("15:04", s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func parseWeekday(s string) time.Weekday {
	switch strings.ToLower(s) {
	case "sun":
		return time.Sunday
	case "mon":
		return time.Monday
	case "tue":
		return time.Tuesday
	case "wed":
		return time.Wednesday
	case "thu":
		return time.Thursday
	case "fri":
		return time.Friday
	case "sat":
		return time.Saturday
	default:
		return time.Sunday // fallback
	}
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
