package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
)

func init() {
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a job",
		Args:  cobra.ExactArgs(1),
		RunE:  runRemove,
	}
	cmd.Flags().Bool("force", false, "skip confirmation")
	cmd.Flags().Bool("with-logs", false, "also remove log files")
	rootCmd.AddCommand(cmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(name)
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return fmt.Errorf("refusing to remove %q without --force (stdin is not a terminal)", name)
		}
		fmt.Printf("Remove job %q? [y/N] ", name)
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if job.Enabled {
		launchd.Bootout(name)
		launchd.RemovePlist(name)
	}

	genPath := launchd.GeneratedPlistPath(s.Dir(), name)
	os.Remove(genPath)

	withLogs, _ := cmd.Flags().GetBool("with-logs")
	if withLogs {
		os.Remove(filepath.Join(s.Dir(), "logs", name+".stdout.log"))
		os.Remove(filepath.Join(s.Dir(), "logs", name+".stderr.log"))
	}

	if err := s.Remove(name); err != nil {
		return err
	}

	if jsonEnabled() {
		writeJSON(os.Stdout, map[string]any{"status": "ok", "name": name})
		return nil
	}
	fmt.Printf("Job %q removed.\n", name)
	return nil
}
