package launchd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"howett.net/plist"

	cronpkg "github.com/dsaiztc/croni/internal/cron"
	"github.com/dsaiztc/croni/internal/types"
)

const labelPrefix = "com.croni."

var validName = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func ValidateName(name string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("job name %q contains invalid characters (allowed: a-z A-Z 0-9 . _ -)", name)
	}
	return nil
}

func Label(name string) string {
	return labelPrefix + name
}

func PlistPath(name string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", Label(name)+".plist"), nil
}

func GeneratedPlistPath(croniDir, name string) string {
	return filepath.Join(croniDir, "generated", Label(name)+".plist")
}

const defaultPath = "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"

func GeneratePlist(job types.Job, croniDir string) ([]byte, error) {
	croniExe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("get executable path: %w", err)
	}

	label := Label(job.Name)
	stdoutLog := filepath.Join(croniDir, "logs", job.Name+".stdout.log")
	stderrLog := filepath.Join(croniDir, "logs", job.Name+".stderr.log")

	env := map[string]string{"PATH": defaultPath}
	for k, v := range job.Env {
		env[k] = v
	}

	pl := map[string]interface{}{
		"Label":             label,
		"ProgramArguments":  []string{croniExe, "_exec", "--job", job.Name},
		"StandardOutPath":   stdoutLog,
		"StandardErrorPath": stderrLog,
		"RunAtLoad":         job.RunOnLoad,
		"EnvironmentVariables": env,
	}

	if job.Workdir != "" {
		pl["WorkingDirectory"] = job.Workdir
	}

	switch job.Schedule.Type {
	case types.ScheduleCron:
		intervals, err := cronpkg.Expand(job.Schedule.Expression)
		if err != nil {
			return nil, fmt.Errorf("expand cron: %w", err)
		}
		pl["StartCalendarInterval"] = intervals

	case types.ScheduleEvery:
		secs, err := ParseInterval(job.Schedule.Expression)
		if err != nil {
			return nil, fmt.Errorf("parse interval: %w", err)
		}
		pl["StartInterval"] = secs

	case types.ScheduleAt:
		t, err := ParseAt(job.Schedule.Expression)
		if err != nil {
			return nil, fmt.Errorf("parse at: %w", err)
		}
		ci := cronpkg.CalendarInterval{}
		minute := t.Minute()
		hour := t.Hour()
		day := t.Day()
		month := int(t.Month())
		ci.Minute = &minute
		ci.Hour = &hour
		ci.Day = &day
		ci.Month = &month
		pl["StartCalendarInterval"] = []cronpkg.CalendarInterval{ci}
	}

	data, err := plist.MarshalIndent(pl, plist.XMLFormat, "\t")
	if err != nil {
		return nil, fmt.Errorf("marshal plist: %w", err)
	}
	return data, nil
}

func ParseInterval(expr string) (int, error) {
	expr = strings.TrimSpace(expr)
	if len(expr) < 2 {
		return 0, fmt.Errorf("invalid interval %q", expr)
	}
	unit := expr[len(expr)-1]
	numStr := expr[:len(expr)-1]
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		return 0, fmt.Errorf("invalid interval %q", expr)
	}
	switch unit {
	case 's':
		return num, nil
	case 'm':
		return num * 60, nil
	case 'h':
		return num * 3600, nil
	case 'd':
		return num * 86400, nil
	default:
		return 0, fmt.Errorf("unknown interval unit %q (use s, m, h, d)", string(unit))
	}
}

func ParseAt(expr string) (time.Time, error) {
	expr = strings.TrimSpace(expr)

	// Relative: "20m", "2h"
	if len(expr) >= 2 {
		unit := expr[len(expr)-1]
		if unit == 'm' || unit == 'h' || unit == 's' {
			numStr := expr[:len(expr)-1]
			if num, err := strconv.Atoi(numStr); err == nil && num > 0 {
				now := time.Now()
				switch unit {
				case 's':
					return now.Add(time.Duration(num) * time.Second), nil
				case 'm':
					return now.Add(time.Duration(num) * time.Minute), nil
				case 'h':
					return now.Add(time.Duration(num) * time.Hour), nil
				}
			}
		}
	}

	// ISO 8601
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, expr, time.Now().Location()); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse time %q (use ISO 8601 or relative like 20m, 2h)", expr)
}
