package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/store"
)

func init() {
	cmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Force-run a job now (synchronous)",
		Args:  cobra.ExactArgs(1),
		RunE:  runRun,
	}
	rootCmd.AddCommand(cmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Running %q...\n", job.Name)

	c := exec.Command("/bin/bash", "-c", job.Command)
	c.Dir = job.Workdir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Env = os.Environ()
	for k, v := range job.Env {
		c.Env = append(c.Env, k+"="+v)
	}

	runErr := c.Run()
	if jsonEnabled() {
		result := map[string]any{"status": "ok", "name": job.Name}
		if runErr != nil {
			result["status"] = "error"
			result["error"] = runErr.Error()
		}
		writeJSON(os.Stdout, result)
		if runErr != nil {
			os.Exit(1)
		}
		return nil
	}
	if runErr != nil {
		return fmt.Errorf("command failed: %w", runErr)
	}
	fmt.Printf("\nJob %q completed successfully.\n", job.Name)
	return nil
}
