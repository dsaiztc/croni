package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	cronpkg "github.com/dsaiztc/croni/internal/cron"
	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
	"github.com/dsaiztc/croni/internal/types"
)

func init() {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Edit an existing job",
		Args:  cobra.ExactArgs(1),
		RunE:  runEdit,
	}
	cmd.Flags().String("cron", "", "5-field cron expression")
	cmd.Flags().String("every", "", "interval shorthand")
	cmd.Flags().String("at", "", "one-shot time")
	cmd.Flags().String("command", "", "command to run")
	cmd.Flags().String("workdir", "", "working directory")
	cmd.Flags().String("description", "", "human-readable description")
	rootCmd.AddCommand(cmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(args[0])
	if err != nil {
		return err
	}

	changed := false

	if cmd.Flags().Changed("command") {
		job.Command, _ = cmd.Flags().GetString("command")
		changed = true
	}
	if cmd.Flags().Changed("workdir") {
		job.Workdir, _ = cmd.Flags().GetString("workdir")
		changed = true
	}
	if cmd.Flags().Changed("description") {
		job.Description, _ = cmd.Flags().GetString("description")
		changed = true
	}

	cronExpr, _ := cmd.Flags().GetString("cron")
	every, _ := cmd.Flags().GetString("every")
	at, _ := cmd.Flags().GetString("at")

	schedCount := 0
	if cmd.Flags().Changed("cron") {
		schedCount++
	}
	if cmd.Flags().Changed("every") {
		schedCount++
	}
	if cmd.Flags().Changed("at") {
		schedCount++
	}
	if schedCount > 1 {
		return fmt.Errorf("only one of --cron, --every, or --at may be specified")
	}

	if cmd.Flags().Changed("cron") {
		if _, err := cronpkg.Expand(cronExpr); err != nil {
			return err
		}
		job.Schedule = types.Schedule{Type: types.ScheduleCron, Expression: cronExpr}
		changed = true
	} else if cmd.Flags().Changed("every") {
		if _, err := launchd.ParseInterval(every); err != nil {
			return err
		}
		job.Schedule = types.Schedule{Type: types.ScheduleEvery, Expression: every}
		changed = true
	} else if cmd.Flags().Changed("at") {
		resolved, err := launchd.ParseAt(at)
		if err != nil {
			return err
		}
		job.Schedule = types.Schedule{Type: types.ScheduleAt, Expression: resolved.Format(time.RFC3339)}
		changed = true
	}

	if !changed {
		return fmt.Errorf("no changes specified")
	}

	job.UpdatedAt = time.Now().UTC()

	if err := s.Update(job); err != nil {
		return err
	}

	if job.Enabled {
		if err := reinstallJob(job, s.Dir()); err != nil {
			return err
		}
	}

	if jsonEnabled() {
		writeJSON(os.Stdout, map[string]any{"status": "ok", "job": job})
		return nil
	}
	fmt.Printf("Job %q updated.\n", job.Name)
	return nil
}

func reinstallJob(job types.Job, croniDir string) error {
	launchd.Bootout(job.Name)

	plistData, err := launchd.GeneratePlist(job, croniDir)
	if err != nil {
		return fmt.Errorf("generate plist: %w", err)
	}

	if err := launchd.InstallPlist(plistData, job.Name); err != nil {
		return fmt.Errorf("install plist: %w", err)
	}

	genPath := launchd.GeneratedPlistPath(croniDir, job.Name)
	if err := writeFile(genPath, plistData); err != nil {
		return err
	}

	plistPath, _ := launchd.PlistPath(job.Name)
	return launchd.Bootstrap(plistPath)
}

func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
