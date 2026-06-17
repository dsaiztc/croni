package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
	"github.com/dsaiztc/croni/internal/types"
)

func init() {
	cmd := &cobra.Command{
		Use:    "_exec",
		Short:  "Internal: execute a job (called by launchd)",
		Hidden: true,
		RunE:   runExec,
	}
	cmd.Flags().String("job", "", "job name")
	cmd.MarkFlagRequired("job")
	rootCmd.AddCommand(cmd)
}

func runExec(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("job")

	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(name)
	if err != nil {
		return err
	}

	c := exec.Command("/bin/bash", "-c", job.Command)
	c.Dir = job.Workdir
	c.Env = os.Environ()
	for k, v := range job.Env {
		c.Env = append(c.Env, k+"="+v)
	}

	stdoutPath := filepath.Join(s.Dir(), "logs", name+".stdout.log")
	stderrPath := filepath.Join(s.Dir(), "logs", name+".stderr.log")

	stdout, err := os.OpenFile(stdoutPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer stdout.Close()

	stderr, err := os.OpenFile(stderrPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer stderr.Close()

	c.Stdout = stdout
	c.Stderr = stderr

	exitCode := 0
	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	if job.Schedule.Type == types.ScheduleAt {
		launchd.Bootout(name)
		launchd.RemovePlist(name)
		s.Remove(name)
	}

	os.Exit(exitCode)
	return nil
}
