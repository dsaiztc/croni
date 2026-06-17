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
		Use:   "disable <name>",
		Short: "Disable an enabled job",
		Args:  cobra.ExactArgs(1),
		RunE:  runDisable,
	}
	rootCmd.AddCommand(cmd)
}

func runDisable(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(args[0])
	if err != nil {
		return err
	}

	if !job.Enabled {
		fmt.Printf("Job %q is already disabled.\n", job.Name)
		return nil
	}

	if err := launchd.Bootout(job.Name); err != nil {
		return fmt.Errorf("unload job: %w", err)
	}
	if err := launchd.RemovePlist(job.Name); err != nil {
		return fmt.Errorf("remove plist: %w", err)
	}

	job.Enabled = false
	job.UpdatedAt = time.Now().UTC()
	if err := s.Update(job); err != nil {
		return err
	}

	if jsonEnabled() {
		writeJSON(os.Stdout, map[string]any{"status": "ok", "job": job})
		return nil
	}
	fmt.Printf("Job %q disabled.\n", job.Name)
	return nil
}
