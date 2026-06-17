package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/store"
)

func init() {
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "View job logs",
		Args:  cobra.ExactArgs(1),
		RunE:  runLogs,
	}
	cmd.Flags().Bool("stderr", false, "show stderr log instead of stdout")
	cmd.Flags().Int("tail", 0, "show last N lines")
	cmd.Flags().BoolP("follow", "f", false, "follow log output")
	rootCmd.AddCommand(cmd)
}

func runLogs(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	if _, err := s.Get(args[0]); err != nil {
		return err
	}

	suffix := ".stdout.log"
	if stderr, _ := cmd.Flags().GetBool("stderr"); stderr {
		suffix = ".stderr.log"
	}
	logPath := filepath.Join(s.Dir(), "logs", args[0]+suffix)

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		fmt.Println("No logs yet.")
		return nil
	}

	follow, _ := cmd.Flags().GetBool("follow")
	tail, _ := cmd.Flags().GetInt("tail")

	if follow {
		var tailArgs []string
		if tail > 0 {
			tailArgs = append(tailArgs, "-n", fmt.Sprintf("%d", tail))
		}
		tailArgs = append(tailArgs, "-f", logPath)
		c := exec.Command("tail", tailArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	if tail > 0 {
		c := exec.Command("tail", "-n", fmt.Sprintf("%d", tail), logPath)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
