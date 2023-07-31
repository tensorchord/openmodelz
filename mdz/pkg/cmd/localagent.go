package cmd

import (
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/mdz/pkg/agentd/server"
)

var (
	localAgentPort int
)

// localAgentCmd represents the local-agent command
var localAgentCmd = &cobra.Command{
	Use:     "local-agent",
	Short:   "Start agent with local docker runtime",
	Long:    `Start agent with local docker runtime`,
	Example: `  mdz local-agent`,
	GroupID: "basic",
	PreRunE: commandInit,
	RunE:    commandLocalAgent,
	Hidden:  true,
}

func init() {
	rootCmd.AddCommand(localAgentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	localAgentCmd.Flags().IntVarP(&localAgentPort, "port", "p", 31112, "Port to listen on")
}

func commandLocalAgent(cmd *cobra.Command, args []string) error {
	server, err := server.New()
	if err != nil {
		return err
	}

	return server.Run(localAgentPort)
}
