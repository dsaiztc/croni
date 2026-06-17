package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

var rootCmd = &cobra.Command{
	Use:   "croni",
	Short: "A friendly CLI for managing scheduled jobs on macOS",
	Long:  "croni abstracts macOS launchd behind a cron-like interface for scheduling recurring commands.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if jsonEnabled() {
			writeJSON(os.Stderr, map[string]any{
				"status": "error",
				"error":  err.Error(),
			})
		}
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = version
	rootCmd.PersistentFlags().Bool("json", false, "output as JSON")
}

func jsonEnabled() bool {
	v, _ := rootCmd.PersistentFlags().GetBool("json")
	return v
}

func writeJSON(f *os.File, v any) {
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
