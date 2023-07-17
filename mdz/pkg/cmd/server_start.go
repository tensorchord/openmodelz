package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/server"
)

// serverStartCmd represents the server start command
var serverStartCmd = &cobra.Command{
	Use:     "start",
	Short:   "Start OpenModelZ server",
	Long:    `Start OpenModelZ server`,
	Example: `  mdz server start`,
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
		return err
	}

	result, err := engine.Run()
	if err != nil {
		return err
	}
	cmd.Printf("🎉 You could set the environment variable to get started!\n")
	cmd.Printf("export MDZ_AGENT=%s\n", result.AgentURL)
	return nil
}
