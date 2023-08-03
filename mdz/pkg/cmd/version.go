package cmd

import (
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/mdz/pkg/version"
)

// versionCmd represents the versionCmd
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print the client and agent version information",
	Long:    `Print the client and server version information`,
	Example: `  mdz version`,
	PreRunE: commandInit,
	RunE:    commandVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandVersion(cmd *cobra.Command, args []string) error {
	v := version.GetVersion()
	cmd.Println("Client:")
	cmd.Printf(" Version: \t%s\n", v.Version)
	cmd.Printf(" Build Date: \t%s\n", v.BuildDate)
	cmd.Printf(" Git Commit: \t%s\n", v.GitCommit)
	cmd.Printf(" Git State: \t%s\n", v.GitTreeState)
	cmd.Printf(" Go Version: \t%s\n", v.GoVersion)
	cmd.Printf(" Compiler: \t%s\n", v.Compiler)
	cmd.Printf(" Platform: \t%s\n", v.Platform)

	if err := printServerVersion(cmd); err != nil {
		cmd.PrintErrf("Failed to get server version: %v\n", errors.Cause(err))
		return err
	}
	return nil
}

func printServerVersion(cmd *cobra.Command) error {
	info, err := agentClient.InfoGet(cmd.Context())
	if err != nil {
		return err
	}

	cmd.Println("Server:")
	cmd.Printf(" Version: \t%s\n", info.Version.Version)
	cmd.Printf(" Build Date: \t%s\n", info.Version.BuildDate)
	cmd.Printf(" Git Commit: \t%s\n", info.Version.GitCommit)
	cmd.Printf(" Git State: \t%s\n", info.Version.GitTreeState)
	cmd.Printf(" Go Version: \t%s\n", info.Version.GoVersion)
	cmd.Printf(" Compiler: \t%s\n", info.Version.Compiler)
	cmd.Printf(" Platform: \t%s\n", info.Version.Platform)
	return nil
}
