package cmd

import (
	"math/rand"
	"time"

	"github.com/cockroachdb/errors"
	petname "github.com/dustinkirkland/golang-petname"
	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

var (
	// Used for flags.
	deployImage       string
	deployPort        int32
	deployMinReplicas int32
	deployMaxReplicas int32
	deployName        string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy OpenModelz inferences",
	Long:  `Deploys OpenModelZ inferences directly via flags.`,
	Example: `  mdz deploy --image=modelzai/llm-blomdz-560m:23.06.13
  mdz deploy --image=modelzai/llm-blomdz-560m:23.06.13 --name blomdz-560m`,
	GroupID: "basic",
	PreRunE: commandInit,
	RunE:    commandDeploy,
}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	rootCmd.AddCommand(deployCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deployCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	deployCmd.Flags().StringVar(&deployImage, "image", "", "Image to deploy")
	deployCmd.Flags().Int32Var(&deployPort, "port", 8080, "Port to deploy on")
	deployCmd.Flags().Int32Var(&deployMinReplicas, "min-replicas", 1, "Minimum number of replicas (can be 0)")
	deployCmd.Flags().Int32Var(&deployMaxReplicas, "max-replicas", 1, "Maximum number of replicas")
	deployCmd.Flags().StringVar(&deployName, "name", "", "Name of inference")
}

func commandDeploy(cmd *cobra.Command, args []string) error {
	if deployImage == "" {
		return cmd.Help()
	}

	name := deployName
	if name == "" {
		name = petname.Generate(2, "-")
	}

	var typ types.ScalingType = types.ScalingTypeCapacity
	inf := types.InferenceDeployment{
		Spec: types.InferenceDeploymentSpec{
			Image:     deployImage,
			Namespace: namespace,
			Name:      name,
			Labels: map[string]string{
				"ai.tensorchord.name": name,
			},
			Framework: types.FrameworkOther,
			Scaling: &types.ScalingConfig{
				MinReplicas:     int32Ptr(deployMinReplicas),
				MaxReplicas:     int32Ptr(deployMaxReplicas),
				TargetLoad:      int32Ptr(10),
				Type:            &typ,
				StartupDuration: int32Ptr(600),
				ZeroDuration:    int32Ptr(600),
			},
			Port: int32Ptr(deployPort),
		},
	}

	if _, err := agentClient.InferenceCreate(
		cmd.Context(), namespace, inf); err != nil {
		cmd.PrintErrf("Failed to create the inference: %s\n", errors.Cause(err))
		return err
	}

	cmd.Printf("Inference %s is created\n", inf.Spec.Name)
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
