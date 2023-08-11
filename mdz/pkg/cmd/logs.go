package cmd

import (
	"github.com/spf13/cobra"
)

var (
	// Used for flags.
	tail   int
	since  string
	end    string
	follow bool
)

// logCmd represents the log command
var logsCmd = &cobra.Command{
	Use:     "logs",
	Short:   "Print the logs for a deployment",
	Long:    `Print the logs for a deployment`,
	Example: `  mdz logs blomdz-560m`,
	GroupID: "debug",
	PreRunE: commandInit,
	Args:    cobra.ExactArgs(1),
	RunE:    commandLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	logsCmd.Flags().IntVarP(&tail, "tail", "t", 0, "Number of lines to show from the end of the logs")
	logsCmd.Flags().StringVarP(&since, "since", "s", "2006-01-02T15:04:05Z", "Show logs since timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")
	logsCmd.Flags().StringVarP(&end, "end", "e", "", "Only return logs before this timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
}

func commandLogs(cmd *cobra.Command, args []string) error {
	logStream, err := agentClient.DeploymentLogGet(cmd.Context(), namespace, args[0], since, tail, end, follow)
	if err != nil {
		cmd.PrintErrf("Failed to get logs: %s\n", err)
		return err
	}
	for log := range logStream {
		cmd.Printf("%s: %s\n", log.Instance, log.Text)
	}
	return nil
}
