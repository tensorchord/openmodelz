package cmd

import "github.com/spf13/cobra"

var serverDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a node from the cluster",
	Long:    `Delete a node from the cluster`,
	Example: `  mdz server delete gpu-node-1`,
	PreRunE: commandInit,
	Args:    cobra.MinimumNArgs(1),
	RunE:    commandServerDelete,
}

func init() {
	serverCmd.AddCommand(serverDeleteCmd)
}

func commandServerDelete(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	if err := agentClient.ServerNodeDelete(cmd.Context(), nodeName); err != nil {
		return err
	}
	return nil
}
