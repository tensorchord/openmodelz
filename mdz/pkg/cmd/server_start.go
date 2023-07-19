package cmd

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/server"
)

// serverStartCmd represents the server start command
var serverStartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start the server",
	Long:    `Start the server`,
	Example: `  mdz server start`,
	PreRunE: commandInitLog,
	RunE:    commandServerStart,
}

func init() {
	serverCmd.AddCommand(serverStartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandServerStart(cmd *cobra.Command, args []string) error {
	engine, err := server.NewStart(server.Options{
		Verbose:       serverVerbose,
		OutputStream:  cmd.ErrOrStderr(),
		RetryInternal: serverPollingInterval,
	})
	if err != nil {
		cmd.PrintErrf("Failed to start the server: %s\n", errors.Cause(err))
		return err
	}

	result, err := engine.Run()
	if err != nil {
		cmd.PrintErrf("Failed to start the server: %s\n", errors.Cause(err))
		return err
	}
	agentURL = result.AgentURL
	if err := commandInit(cmd, args); err != nil {
		cmd.PrintErrf("Failed to start the server: %s\n", errors.Cause(err))
		return err
	}

	cmd.Printf("🐋 Checking if the server is running...\n")
	// Retry until verify success.
	ticker := time.NewTicker(serverPollingInterval)
	for range ticker.C {
		if err := printAgentVersion(cmd); err != nil {
			cmd.Printf("🐋 The server is not ready yet, retrying...\n")
			continue
		}
		break
	}
	cmd.Printf("🐳 The server is running at %s\n", result.AgentURL)
	cmd.Printf("🎉 You could set the environment variable to get started!\n\n")
	cmd.Printf("export MDZ_AGENT=%s\n", result.AgentURL)
	return nil
}
