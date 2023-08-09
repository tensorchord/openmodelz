package cmd

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/server"
)

// serverDestroyCmd represents the server destroy command
var serverDestroyCmd = &cobra.Command{
	Use:     "destroy",
	Short:   "Destroy the cluster",
	Long:    `Destroy the cluster`,
	Example: `  mdz server destroy`,
	PreRunE: commandInitLog,
	RunE:    commandServerDestroy,
}

func init() {
	serverCmd.AddCommand(serverDestroyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandServerDestroy(cmd *cobra.Command, args []string) error {
	engine, err := server.NewDestroy(server.Options{
		Verbose:       serverVerbose,
		OutputStream:  cmd.ErrOrStderr(),
		RetryInternal: serverPollingInterval,
	})
	if err != nil {
		cmd.PrintErrf("Failed to destroy the server: %s\n", errors.Cause(err))
		return err
	}

	_, err = engine.Run()
	if err != nil {
		cmd.PrintErrf("Failed to destroy the server: %s\n", errors.Cause(err))
		return err
	}
	cmd.Printf("✅ Server destroyed\n")
	return nil
}
