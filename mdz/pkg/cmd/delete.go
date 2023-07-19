package cmd

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete OpenModelz inferences",
	Long:    `Deletes OpenModelZ inferences`,
	Example: `  mdz delete blomdz-560m`,
	GroupID: "basic",
	PreRunE: commandInit,
	Args:    cobra.ExactArgs(1),
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
	name := args[0]

	if err := agentClient.InferenceRemove(
		cmd.Context(), namespace, name); err != nil {
		cmd.PrintErrf("Failed to remove the inference: %s\n", errors.Cause(err))
		return err
	}

	cmd.Printf("Inference %s is deleted\n", name)
	return nil
}
