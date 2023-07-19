package cmd

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/server"
)

// serverJoinCmd represents the server join command
var serverJoinCmd = &cobra.Command{
	Use:     "join",
	Short:   "Join to the cluster",
	Long:    `Join to the cluster`,
	Example: `  mdz server join 192.168.31.192`,
	PreRunE: commandInitLog,
	Args:    cobra.ExactArgs(1),
	RunE:    commandServerJoin,
}

func init() {
	serverCmd.AddCommand(serverJoinCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandServerJoin(cmd *cobra.Command, args []string) error {
	engine, err := server.NewJoin(server.Options{
		Verbose:       serverVerbose,
		OutputStream:  cmd.ErrOrStderr(),
		RetryInternal: serverPollingInterval,
		ServerIP:      args[0],
	})
	if err != nil {
		cmd.PrintErrf("Failed to join the cluster: %s\n", errors.Cause(err))
		return err
	}

	_, err = engine.Run()
	if err != nil {
		cmd.PrintErrf("Failed to join the cluster: %s\n", errors.Cause(err))
		return err
	}
	cmd.Printf("✅ Server Joined\n")
	return nil
}
