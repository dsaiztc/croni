package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
)

func init() {
	cmd := &cobra.Command{
		Use:   "enable <name>",
		Short: "Enable a disabled job",
		Args:  cobra.ExactArgs(1),
		RunE:  runEnable,
	}
	rootCmd.AddCommand(cmd)
}

func runEnable(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(args[0])
	if err != nil {
		return err
	}

	if job.Enabled {
		fmt.Printf("Job %q is already enabled.\n", job.Name)
		return nil
	}

	job.Enabled = true
	job.UpdatedAt = time.Now().UTC()
	if err := s.Update(job); err != nil {
		return err
	}

	plistData, err := launchd.GeneratePlist(job, s.Dir())
	if err != nil {
		return fmt.Errorf("generate plist: %w", err)
	}

	if err := launchd.InstallPlist(plistData, job.Name); err != nil {
		return fmt.Errorf("install plist: %w", err)
	}

	plistPath, _ := launchd.PlistPath(job.Name)
	if err := launchd.Bootstrap(plistPath); err != nil {
		return fmt.Errorf("load job: %w", err)
	}

	if jsonEnabled() {
		writeJSON(os.Stdout, map[string]any{"status": "ok", "job": job})
		return nil
	}
	fmt.Printf("Job %q enabled.\n", job.Name)
	return nil
}
