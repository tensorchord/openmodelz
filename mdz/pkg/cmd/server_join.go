package cmd

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/server"
	"github.com/tensorchord/openmodelz/mdz/pkg/telemetry"
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
	serverJoinCmd.Flags().StringVarP(&serverRegistryMirrorName, "mirror-name", "",
		"docker.io", "Mirror domain name of the registry")
	serverJoinCmd.Flags().StringArrayVarP(&serverRegistryMirrorEndpoints, "mirror-endpoints", "",
		[]string{}, "Mirror URL endpoints of the registry like `https://quay.io`")
}

func commandServerJoin(cmd *cobra.Command, args []string) error {
	engine, err := server.NewJoin(server.Options{
		Verbose:       serverVerbose,
		OutputStream:  cmd.ErrOrStderr(),
		RetryInternal: serverPollingInterval,
		ServerIP:      args[0],
		Mirror: server.Mirror{
			Name:      serverRegistryMirrorName,
			Endpoints: serverRegistryMirrorEndpoints,
		},
	})
	if err != nil {
		cmd.PrintErrf("Failed to configure before join: %s\n", errors.Cause(err))
		return err
	}

	telemetry.GetTelemetry().Record("server join")

	_, err = engine.Run()
	if err != nil {
		cmd.PrintErrf("Failed to join the cluster: %s\n", errors.Cause(err))
		return err
	}
	cmd.Printf("✅ Server joined\n")
	return nil
}
