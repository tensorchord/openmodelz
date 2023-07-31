package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// serverLabelCmd represents the server label command
var serverLabelCmd = &cobra.Command{
	Use:   "label",
	Short: "Update the labels on a server",
	Long: `Update the labels on a server

  *  A label key and value must begin with a letter or number, and may contain letters, numbers, hyphens, dots, and underscores, up to 63 characters each.
  *  Optionally, the key can begin with a DNS subdomain prefix and a single '/', like example.com/my-app.
	`,
	Example: `  mdz server label node-name key=value [key=value...]`,
	PreRunE: commandInit,
	Args:    cobra.MinimumNArgs(1),
	RunE:    commandServerLabel,
}

func init() {
	serverCmd.AddCommand(serverLabelCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandServerLabel(cmd *cobra.Command, args []string) error {
	nodeName := args[0]
	labels := args[1:]

	nodeLabels, err := parseNodeLabels(labels)
	if err != nil {
		return err
	}

	if err := agentClient.ServerLabelCreate(cmd.Context(),
		nodeName, nodeLabels); err != nil {
		return err
	}

	return nil
}

func parseNodeLabels(labels []string) (map[string]string, error) {
	res := make(map[string]string)
	for _, label := range labels {
		if !strings.Contains(label, "=") {
			return nil, fmt.Errorf("label must be in the form of key=value")
		}
		// Split the label into key and value
		parts := strings.SplitN(label, "=", 2)
		key := parts[0]
		value := parts[1]
		if len(key) == 0 {
			return nil, fmt.Errorf("label key cannot be empty")
		}
		res[key] = value
	}
	return res, nil
}
