package cmd

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/server"
)

// serverStopCmd represents the server stop command
var serverStopCmd = &cobra.Command{
	Use:     "stop",
	Short:   "Stop the server",
	Long:    `Stop the server`,
	Example: `  mdz server stop`,
	PreRunE: commandInitLog,
	RunE:    commandServerStop,
}

func init() {
	serverCmd.AddCommand(serverStopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandServerStop(cmd *cobra.Command, args []string) error {
	engine, err := server.NewStop(server.Options{
		Verbose:       serverVerbose,
		OutputStream:  cmd.ErrOrStderr(),
		RetryInternal: serverPollingInterval,
	})
	if err != nil {
		cmd.PrintErrf("Failed to stop the server: %s\n", errors.Cause(err))
		return err
	}

	_, err = engine.Run()
	if err != nil {
		cmd.PrintErrf("Failed to stop the server: %s\n", errors.Cause(err))
		return err
	}
	cmd.Printf("✅ Server stopped\n")
	return nil
}
