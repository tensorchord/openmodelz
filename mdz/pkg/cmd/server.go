package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var (
	serverVerbose         bool
	serverPollingInterval time.Duration
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Manage OpenModelZ servers",
	Long:    `Manage OpenModelZ servers`,
	Example: `  mdz server start`,
	GroupID: "management",
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	serverCmd.PersistentFlags().BoolVarP(&serverVerbose, "verbose", "v", false, "Verbose output")
	serverCmd.PersistentFlags().DurationVarP(&serverPollingInterval, "polling-interval", "p", 5*time.Second, "Polling interval")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
