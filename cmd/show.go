package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/store"
)

func init() {
	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Show details of a job",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}
	rootCmd.AddCommand(cmd)
}

func runShow(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(args[0])
	if err != nil {
		return err
	}

	if jsonEnabled() {
		writeJSON(os.Stdout, job)
		return nil
	}

	fmt.Printf("Name:        %s\n", job.Name)
	fmt.Printf("Command:     %s\n", job.Command)
	fmt.Printf("Workdir:     %s\n", job.Workdir)
	fmt.Printf("Schedule:    %s %s\n", job.Schedule.Type, job.Schedule.Expression)
	if job.Description != "" {
		fmt.Printf("Description: %s\n", job.Description)
	}
	fmt.Printf("Enabled:     %t\n", job.Enabled)
	fmt.Printf("RunOnLoad:   %t\n", job.RunOnLoad)
	fmt.Printf("Created:     %s\n", job.CreatedAt.Local().Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated:     %s\n", job.UpdatedAt.Local().Format("2006-01-02 15:04:05"))
	if len(job.Env) > 0 {
		fmt.Println("Env:")
		for k, v := range job.Env {
			fmt.Printf("  %s=%s\n", k, v)
		}
	}
	nextRun := computeNextRun(job)
	fmt.Printf("Next run:    %s\n", nextRun)
	return nil
}
