package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/tensorchord/openmodelz/agent/client"
)

var (
	// Used for flags.
	agentURL  string
	namespace string
	debug     bool

	agentClient *client.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mdz",
	Short: "Manage your OpenModelZ inferences from the command line",
	Long:  `mdz is a CLI library to manage your OpenModelZ inferences from the command line.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mdz.yaml)")
	rootCmd.PersistentFlags().StringVarP(&agentURL, "agent-url", "a", "http://localhost:8081", "URL of the OpenModelZ agent")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to use for OpenModelZ inferences")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddGroup(&cobra.Group{ID: "basic", Title: "Basic Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "debug", Title: "Troubleshooting and Debugging Commands:"})
}

func getAgentClient(cmd *cobra.Command, args []string) error {
	if agentClient == nil {
		var err error
		agentClient, err = client.NewClientWithOpts(client.WithHost(agentURL))
		if err != nil {
			return err
		}
	}
	return nil
}

func GenMarkdownTree(dir string) error {
	return doc.GenMarkdownTree(rootCmd, dir)
}
