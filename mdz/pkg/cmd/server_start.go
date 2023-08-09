package cmd

import (
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/tensorchord/openmodelz/agent/pkg/consts"
	"github.com/tensorchord/openmodelz/mdz/pkg/server"
	"github.com/tensorchord/openmodelz/mdz/pkg/telemetry"
	"github.com/tensorchord/openmodelz/mdz/pkg/version"
)

var (
	serverStartRuntime string
	serverStartDomain  string = consts.Domain
	serverStartVersion string
	serverStartWithGPU bool
)

// serverStartCmd represents the server start command
var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the server",
	Long:  `Start the server with the public IP of the machine. If not provided, the internal IP will be used automatically.`,
	Example: `  mdz server start
  mdz server start -v
  mdz server start 1.2.3.4`,
	PreRunE: commandInitLog,
	Args:    cobra.RangeArgs(0, 1),
	RunE:    commandServerStart,
}

func init() {
	serverCmd.AddCommand(serverStartCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverStartCmd.Flags().StringVarP(&serverStartRuntime, "runtime", "r", "k3s", "Runtime to use (k3s, docker) in the started server")
	serverStartCmd.Flags().StringVarP(&serverStartVersion, "version", "",
		version.HelmChartVersion, "Version of the server to start")
	serverStartCmd.Flags().MarkHidden("version")
	serverStartCmd.Flags().BoolVarP(&serverStartWithGPU, "force-gpu", "g",
		false, "Start the server with GPU support (ignore the GPU detection)")
	serverStartCmd.Flags().StringVarP(&serverRegistryMirrorName, "mirror-name", "",
		"docker.io", "Mirror domain name of the registry")
	serverStartCmd.Flags().StringArrayVarP(&serverRegistryMirrorEndpoints, "mirror-endpoints", "",
		[]string{}, "Mirror URL endpoints of the registry like `https://quay.io`")
}

func commandServerStart(cmd *cobra.Command, args []string) error {
	var domain *string
	if len(args) > 0 {
		domainWithSuffix := fmt.Sprintf("%s.%s", args[0], serverStartDomain)
		domain = &domainWithSuffix
	}
	defer func(start time.Time) {
		telemetry.GetTelemetry().Record(
			"server start",
			telemetry.AddField("duration", time.Since(start).Seconds()),
		)
	}(time.Now())
	engine, err := server.NewStart(server.Options{
		Verbose:       serverVerbose,
		Runtime:       server.Runtime(serverStartRuntime),
		OutputStream:  cmd.ErrOrStderr(),
		RetryInternal: serverPollingInterval,
		Domain:        domain,
		Version:       serverStartVersion,
		ForceGPU:      serverStartWithGPU,
		Mirror: server.Mirror{
			Name:      serverRegistryMirrorName,
			Endpoints: serverRegistryMirrorEndpoints,
		},
	})
	if err != nil {
		cmd.PrintErrf("Failed to start the server: %s\n", errors.Cause(err))
		return err
	}

	result, err := engine.Run()
	if err != nil {
		cmd.PrintErrf("Failed to start the server: %s\n", errors.Cause(err))
		return err
	}
	mdzURL = result.MDZURL
	if err := commandInit(cmd, args); err != nil {
		cmd.PrintErrf("Failed to start the server: %s\n", errors.Cause(err))
		return err
	}

	cmd.Printf("🐋 Checking if the server is running...\n")
	// Retry until verify success.
	ticker := time.NewTicker(serverPollingInterval)
	for range ticker.C {
		if err := printServerVersion(cmd); err != nil {
			cmd.Printf("🐋 The server is not ready yet, retrying...\n")
			continue
		}
		break
	}
	cmd.Printf("🐳 The server is running at %s\n", mdzURL)
	cmd.Printf("🎉 You could set the environment variable to get started!\n\n")
	cmd.Printf("export MDZ_URL=%s\n", mdzURL)
	return nil
}
