package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
	"github.com/dsaiztc/croni/internal/types"
)

func init() {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new scheduled job",
		Args:  cobra.ExactArgs(1),
		RunE:  runAdd,
	}
	cmd.Flags().String("cron", "", "5-field cron expression")
	cmd.Flags().String("every", "", "interval shorthand (e.g. 5m, 2h, 1d)")
	cmd.Flags().String("at", "", "one-shot time (ISO 8601 or relative like 20m)")
	cmd.Flags().String("command", "", "command to run")
	cmd.Flags().String("workdir", "", "working directory (defaults to $PWD)")
	cmd.Flags().StringArray("env", nil, "environment variable KEY=VAL (repeatable)")
	cmd.Flags().String("description", "", "human-readable description")
	cmd.Flags().Bool("disabled", false, "create without loading into launchd")
	cmd.Flags().Bool("run-on-load", false, "run immediately when enabled")
	cmd.MarkFlagRequired("command")
	rootCmd.AddCommand(cmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := launchd.ValidateName(name); err != nil {
		return err
	}

	cronExpr, _ := cmd.Flags().GetString("cron")
	every, _ := cmd.Flags().GetString("every")
	at, _ := cmd.Flags().GetString("at")

	schedCount := 0
	if cronExpr != "" {
		schedCount++
	}
	if every != "" {
		schedCount++
	}
	if at != "" {
		schedCount++
	}
	if schedCount != 1 {
		return fmt.Errorf("exactly one of --cron, --every, or --at is required")
	}

	var sched types.Schedule
	switch {
	case cronExpr != "":
		sched = types.Schedule{Type: types.ScheduleCron, Expression: cronExpr}
	case every != "":
		if _, err := launchd.ParseInterval(every); err != nil {
			return err
		}
		sched = types.Schedule{Type: types.ScheduleEvery, Expression: every}
	case at != "":
		resolved, err := launchd.ParseAt(at)
		if err != nil {
			return err
		}
		// Store as absolute RFC3339 so subsequent ParseAt calls (list, plist
		// regeneration) always return the same fixed time, not time.Now()+offset.
		sched = types.Schedule{Type: types.ScheduleAt, Expression: resolved.Format(time.RFC3339)}
	}

	command, _ := cmd.Flags().GetString("command")
	workdir, _ := cmd.Flags().GetString("workdir")
	if workdir == "" {
		workdir, _ = os.Getwd()
	}

	envSlice, _ := cmd.Flags().GetStringArray("env")
	env := make(map[string]string)
	for _, e := range envSlice {
		for i := range e {
			if e[i] == '=' {
				env[e[:i]] = e[i+1:]
				break
			}
		}
	}

	description, _ := cmd.Flags().GetString("description")
	disabled, _ := cmd.Flags().GetBool("disabled")

	runOnLoad, _ := cmd.Flags().GetBool("run-on-load")
	if !cmd.Flags().Changed("run-on-load") {
		runOnLoad = sched.Type == types.ScheduleEvery
	}

	now := time.Now().UTC()
	job := types.Job{
		Name:        name,
		Command:     command,
		Workdir:     workdir,
		Schedule:    sched,
		Env:         env,
		Description: description,
		RunOnLoad:   runOnLoad,
		Enabled:     !disabled,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s, err := store.New()
	if err != nil {
		return err
	}

	if err := s.Add(job); err != nil {
		return err
	}

	plistData, err := launchd.GeneratePlist(job, s.Dir())
	if err != nil {
		s.Remove(name)
		return fmt.Errorf("generate plist: %w", err)
	}

	genPath := launchd.GeneratedPlistPath(s.Dir(), name)
	if err := os.WriteFile(genPath, plistData, 0644); err != nil {
		s.Remove(name)
		return fmt.Errorf("write generated plist: %w", err)
	}

	if !disabled {
		if err := launchd.InstallPlist(plistData, name); err != nil {
			s.Remove(name)
			return fmt.Errorf("install plist: %w", err)
		}
		plistPath, _ := launchd.PlistPath(name)
		if err := launchd.Bootstrap(plistPath); err != nil {
			launchd.RemovePlist(name)
			s.Remove(name)
			return fmt.Errorf("load job: %w", err)
		}
	}

	if jsonEnabled() {
		writeJSON(os.Stdout, map[string]any{"status": "ok", "job": job})
		return nil
	}
	fmt.Printf("Job %q created", name)
	if disabled {
		fmt.Printf(" (disabled)")
	}
	fmt.Println()
	fmt.Printf("  schedule: %s %s\n", sched.Type, sched.Expression)
	fmt.Printf("  command:  %s\n", command)
	fmt.Printf("  logs:     croni logs %s\n", name)
	return nil
}
