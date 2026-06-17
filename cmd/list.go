package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
	"github.com/dsaiztc/croni/internal/types"
)

func init() {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all jobs",
		Args:  cobra.NoArgs,
		RunE:  runList,
	}
	cmd.Flags().Bool("enabled", false, "show only enabled jobs")
	cmd.Flags().Bool("disabled", false, "show only disabled jobs")
	rootCmd.AddCommand(cmd)
}

func runList(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	jobs, err := s.List()
	if err != nil {
		return err
	}

	enabledOnly, _ := cmd.Flags().GetBool("enabled")
	disabledOnly, _ := cmd.Flags().GetBool("disabled")

	filtered := make([]types.Job, 0)
	for _, j := range jobs {
		if enabledOnly && !j.Enabled {
			continue
		}
		if disabledOnly && j.Enabled {
			continue
		}
		filtered = append(filtered, j)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})

	if jsonEnabled() {
		writeJSON(os.Stdout, filtered)
		return nil
	}

	if len(filtered) == 0 {
		fmt.Println("No jobs found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSCHEDULE\tSTATUS\tNEXT RUN")
	for _, j := range filtered {
		schedule := formatSchedule(j.Schedule)
		status := "enabled"
		if !j.Enabled {
			status = "disabled"
		}
		nextRun := computeNextRun(j)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", j.Name, schedule, status, nextRun)
	}
	w.Flush()
	return nil
}

func formatSchedule(s types.Schedule) string {
	switch s.Type {
	case types.ScheduleCron:
		return s.Expression
	case types.ScheduleEvery:
		return "every " + s.Expression
	case types.ScheduleAt:
		return "at " + s.Expression
	default:
		return s.Expression
	}
}

func computeNextRun(j types.Job) string {
	if !j.Enabled {
		return "-"
	}
	switch j.Schedule.Type {
	case types.ScheduleCron:
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		sched, err := parser.Parse(j.Schedule.Expression)
		if err != nil {
			return "?"
		}
		return sched.Next(time.Now()).Format("2006-01-02 15:04:05")
	case types.ScheduleEvery:
		secs, err := launchd.ParseInterval(j.Schedule.Expression)
		if err != nil {
			return "?"
		}
		return time.Now().Add(time.Duration(secs) * time.Second).Format("2006-01-02 15:04:05")
	case types.ScheduleAt:
		t, err := launchd.ParseAt(j.Schedule.Expression)
		if err != nil {
			return "?"
		}
		if t.Before(time.Now()) {
			return "passed"
		}
		return t.Format("2006-01-02 15:04:05")
	}
	return "?"
}
