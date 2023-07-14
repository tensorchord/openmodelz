package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete OpenModelz inferences",
	Long:    `Deletes OpenModelZ inferences`,
	Example: `omz delete <name>`,
	PreRunE: getAgentClient,
	RunE:    commandDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandDelete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return cmd.Help()
	}

	name := args[0]

	if err := agentClient.InferenceRemove(
		cmd.Context(), namespace, name); err != nil {
		return err
	}

	fmt.Printf("Inference %s is deleted\n", name)
	return nil
}
