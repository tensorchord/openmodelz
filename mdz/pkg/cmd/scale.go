package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/mdz/pkg/telemetry"
)

var (
	// Used for flags.
	replicas               int32
	min                    int32
	max                    int32
	targetInflightRequests int32
)

// scaleCmd represents the scale command
var scaleCmd = &cobra.Command{
	Use:     "scale",
	Short:   "Scale a deployment",
	Long:    `Scale a deployment`,
	Example: `  mdz scale bloomz-560m --replicas 3`,
	GroupID: "basic",
	PreRunE: commandInit,
	Args:    cobra.ExactArgs(1),
	RunE:    commandScale,
}

func init() {
	rootCmd.AddCommand(scaleCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	scaleCmd.Flags().Int32VarP(&replicas, "replicas", "r", 0, "Number of replicas to scale to")
	scaleCmd.Flags().Int32VarP(&min, "min", "m", 0, "Minimum number of replicas to scale to")
	scaleCmd.Flags().Int32VarP(&max, "max", "x", 0, "Maximum number of replicas to scale to")
	scaleCmd.Flags().Int32VarP(&targetInflightRequests, "target-inflight-requests", "t", 0, "Target number of inflight requests per replica")
	scaleCmd.MarkFlagRequired("replicas")
	scaleCmd.Flags().MarkHidden("min")
	scaleCmd.Flags().MarkHidden("max")
	scaleCmd.Flags().MarkHidden("target-inflight-requests")
}

func commandScale(cmd *cobra.Command, args []string) error {
	name := args[0]
	deployment, err := agentClient.InferenceGet(cmd.Context(), namespace, name)
	if err != nil {
		cmd.PrintErrf("Failed to get deployment: %s\n", err)
		return err
	}

	if replicas != 0 {
		deployment.Spec.Scaling.MinReplicas = int32Ptr(replicas)
		deployment.Spec.Scaling.MaxReplicas = int32Ptr(replicas)

		if _, err := agentClient.DeploymentUpdate(cmd.Context(), namespace, deployment); err != nil {
			cmd.PrintErrf("Failed to update deployment: %s\n", err)
			return err
		}
		return nil
	}

	if min != 0 {
		deployment.Spec.Scaling.MinReplicas = int32Ptr(min)
	}
	if max != 0 {
		deployment.Spec.Scaling.MaxReplicas = int32Ptr(max)
	}
	if targetInflightRequests != 0 {
		deployment.Spec.Scaling.TargetLoad = int32Ptr(targetInflightRequests)
	}

	telemetry.GetTelemetry().Record("scale")

	if _, err := agentClient.DeploymentUpdate(cmd.Context(), namespace, deployment); err != nil {
		cmd.PrintErrf("Failed to update deployment: %s\n", err)
		return err
	}

	return nil
}
