package cron

import (
	"fmt"
	"strconv"
	"strings"
)

type CalendarInterval struct {
	Minute  *int `plist:"Minute,omitempty"`
	Hour    *int `plist:"Hour,omitempty"`
	Day     *int `plist:"Day,omitempty"`
	Month   *int `plist:"Month,omitempty"`
	Weekday *int `plist:"Weekday,omitempty"`
}

const maxExpansion = 512

var specialExpressions = map[string]string{
	"@yearly":   "0 0 1 1 *",
	"@annually": "0 0 1 1 *",
	"@monthly":  "0 0 1 * *",
	"@weekly":   "0 0 * * 0",
	"@daily":    "0 0 * * *",
	"@midnight": "0 0 * * *",
	"@hourly":   "0 * * * *",
}

var monthNames = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4,
	"may": 5, "jun": 6, "jul": 7, "aug": 8,
	"sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

var dayNames = map[string]int{
	"sun": 0, "mon": 1, "tue": 2, "wed": 3,
	"thu": 4, "fri": 5, "sat": 6,
}

func Expand(expr string) ([]CalendarInterval, error) {
	expr = strings.TrimSpace(expr)
	if replacement, ok := specialExpressions[strings.ToLower(expr)]; ok {
		expr = replacement
	}

	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cron expression must have exactly 5 fields, got %d", len(fields))
	}

	minutes, err := expandField(fields[0], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("minute field: %w", err)
	}
	hours, err := expandField(fields[1], 0, 23, nil)
	if err != nil {
		return nil, fmt.Errorf("hour field: %w", err)
	}
	days, err := expandField(fields[2], 1, 31, nil)
	if err != nil {
		return nil, fmt.Errorf("day field: %w", err)
	}
	months, err := expandField(fields[3], 1, 12, monthNames)
	if err != nil {
		return nil, fmt.Errorf("month field: %w", err)
	}
	weekdays, err := expandField(fields[4], 0, 6, dayNames)
	if err != nil {
		return nil, fmt.Errorf("weekday field: %w", err)
	}

	hasDays := fields[2] != "*"
	hasWeekdays := fields[4] != "*"

	// launchd treats day+weekday as AND, cron treats as OR.
	// When both are non-wildcard, generate two separate sets and merge.
	if hasDays && hasWeekdays {
		var intervals []CalendarInterval
		dayIntervals := cartesian(minutes, hours, days, months, nil)
		weekdayIntervals := cartesian(minutes, hours, nil, months, weekdays)
		intervals = append(intervals, dayIntervals...)
		intervals = append(intervals, weekdayIntervals...)
		if len(intervals) > maxExpansion {
			return nil, fmt.Errorf("cron expression expands to %d calendar intervals (max %d)", len(intervals), maxExpansion)
		}
		return intervals, nil
	}

	intervals := cartesian(minutes, hours, days, months, weekdays)
	if len(intervals) > maxExpansion {
		return nil, fmt.Errorf("cron expression expands to %d calendar intervals (max %d)", len(intervals), maxExpansion)
	}
	return intervals, nil
}

func cartesian(minutes, hours, days, months, weekdays []int) []CalendarInterval {
	minSet := toOptional(minutes)
	hourSet := toOptional(hours)
	daySet := toOptional(days)
	monthSet := toOptional(months)
	wdaySet := toOptional(weekdays)

	var result []CalendarInterval
	for _, mi := range minSet {
		for _, hi := range hourSet {
			for _, di := range daySet {
				for _, moi := range monthSet {
					for _, wi := range wdaySet {
						ci := CalendarInterval{
							Minute:  mi,
							Hour:    hi,
							Day:     di,
							Month:   moi,
							Weekday: wi,
						}
						result = append(result, ci)
					}
				}
			}
		}
	}
	return result
}

func toOptional(vals []int) []*int {
	if vals == nil {
		return []*int{nil}
	}
	result := make([]*int, len(vals))
	for i := range vals {
		v := vals[i]
		result[i] = &v
	}
	return result
}

func expandField(field string, min, max int, names map[string]int) ([]int, error) {
	if field == "*" {
		return nil, nil
	}

	var result []int
	parts := strings.Split(field, ",")
	for _, part := range parts {
		vals, err := expandPart(part, min, max, names)
		if err != nil {
			return nil, err
		}
		result = append(result, vals...)
	}

	seen := make(map[int]bool)
	unique := result[:0]
	for _, v := range result {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}
	return unique, nil
}

func expandPart(part string, min, max int, names map[string]int) ([]int, error) {
	step := 1
	if idx := strings.Index(part, "/"); idx >= 0 {
		s, err := strconv.Atoi(part[idx+1:])
		if err != nil || s <= 0 {
			return nil, fmt.Errorf("invalid step in %q", part)
		}
		step = s
		part = part[:idx]
	}

	if part == "*" {
		var vals []int
		for i := min; i <= max; i += step {
			vals = append(vals, i)
		}
		return vals, nil
	}

	if idx := strings.Index(part, "-"); idx >= 0 {
		lo, err := parseValue(part[:idx], min, max, names)
		if err != nil {
			return nil, fmt.Errorf("invalid range start in %q: %w", part, err)
		}
		hi, err := parseValue(part[idx+1:], min, max, names)
		if err != nil {
			return nil, fmt.Errorf("invalid range end in %q: %w", part, err)
		}
		if lo > hi {
			return nil, fmt.Errorf("range %d-%d: start exceeds end", lo, hi)
		}
		var vals []int
		for i := lo; i <= hi; i += step {
			vals = append(vals, i)
		}
		return vals, nil
	}

	val, err := parseValue(part, min, max, names)
	if err != nil {
		return nil, err
	}
	return []int{val}, nil
}

func parseValue(s string, min, max int, names map[string]int) (int, error) {
	if names != nil {
		if v, ok := names[strings.ToLower(s)]; ok {
			return v, nil
		}
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q", s)
	}
	if val < min || val > max {
		return 0, fmt.Errorf("value %d out of bounds [%d, %d]", val, min, max)
	}
	return val, nil
}
