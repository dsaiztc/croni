package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dsaiztc/croni/internal/launchd"
	"github.com/dsaiztc/croni/internal/store"
)

func init() {
	cmd := &cobra.Command{
		Use:   "export <name>",
		Short: "Dump generated plist to stdout",
		Args:  cobra.ExactArgs(1),
		RunE:  runExport,
	}
	rootCmd.AddCommand(cmd)
}

func runExport(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return err
	}

	job, err := s.Get(args[0])
	if err != nil {
		return err
	}

	plistData, err := launchd.GeneratePlist(job, s.Dir())
	if err != nil {
		return err
	}

	fmt.Fprint(os.Stdout, string(plistData))
	return nil
}
