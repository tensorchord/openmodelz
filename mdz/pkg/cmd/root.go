package cmd

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/tensorchord/openmodelz/agent/client"
	"github.com/tensorchord/openmodelz/mdz/pkg/telemetry"
)

var (
	// Used for flags.
	mdzURL           string
	namespace        string
	debug            bool
	disableTelemetry bool

	agentClient *client.Client
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mdz",
	Short: "mdz manages your deployments",
	Long:  `mdz helps you deploy applications, manage servers, and troubleshoot issues.`,
	Example: `  mdz server start
  mdz deploy --image modelzai/llm-bloomz-560m:23.06.13 --name llm
  mdz list
  mdz logs llm
  mdz port-forward llm 7860
  mdz exec llm ps
  mdz exec llm --tty bash
  mdz delete llm
`,
	SilenceUsage: true,
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
	rootCmd.PersistentFlags().StringVarP(&mdzURL, "url", "u", "", "URL to use for the server (MDZ_URL) (default http://localhost:80)")

	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "Namespace to use for OpenModelZ inferences")
	rootCmd.PersistentFlags().MarkHidden("namespace")

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logging")

	rootCmd.PersistentFlags().BoolVarP(&disableTelemetry, "disable-telemetry", "", false, "Disable anonymous telemetry")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.AddGroup(&cobra.Group{ID: "basic", Title: "Basic Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "debug", Title: "Troubleshooting and Debugging Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "management", Title: "Management Commands:"})

	// telemetry
	if err := telemetry.Initialize(!disableTelemetry); err != nil {
		logrus.WithError(err).Debug("Failed to initialize telemetry")
	}
}

func commandInit(cmd *cobra.Command, args []string) error {
	if err := commandInitLog(cmd, args); err != nil {
		return err
	}

	if agentClient == nil {
		if mdzURL == "" {
			// Checkout environment variable MDZ_URL.
			mdzURL = os.Getenv("MDZ_URL")
		}
		if mdzURL == "" {
			mdzURL = "http://localhost:80"
		}
		var err error
		agentClient, err = client.NewClientWithOpts(client.WithHost(mdzURL))
		if err != nil {
			cmd.PrintErrf("Failed to connect to agent: %s\n", errors.Cause(err))
			return err
		}
	}

	return nil
}

func commandInitLog(cmd *cobra.Command, args []string) error {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Debug logging enabled")
		logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	}
	return nil
}

func GenMarkdownTree(dir string) error {
	return doc.GenMarkdownTree(rootCmd, dir)
}
