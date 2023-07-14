package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

var (
	// Used for flags.
	listQuiet   bool
	listVerbose bool
)

// listCommand represents the list command
var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List OpenModelz inferences",
	Long:  `Lists OpenModelZ inferences either on a local or remote agent`,
	Example: `  omz list
  omz list -v
  omz list -q`,
	GroupID: "basic",
	PreRunE: getAgentClient,
	RunE:    commandList,
}

func init() {
	rootCmd.AddCommand(listCommand)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listCommand.Flags().BoolVarP(&listQuiet, "quiet", "q", false, "Quiet mode - print out only the inference names")
	listCommand.Flags().BoolVarP(&listVerbose, "verbose", "v", false, "Verbose mode - print out all inference details")
}

func commandList(cmd *cobra.Command, args []string) error {
	infs, err := agentClient.InferenceList(cmd.Context(), namespace)
	if err != nil {
		return err
	}

	sort.Sort(byName(infs))
	if listQuiet {
		for _, inf := range infs {
			cmd.Printf("%s\n", inf.Spec.Name)
		}
		return nil
	} else if listVerbose {
		maxWidth := 40
		for _, inf := range infs {
			if len(inf.Spec.Image) > maxWidth {
				maxWidth = len(inf.Spec.Image)
			}
		}

		cmd.Printf("%-30s\t%-"+fmt.Sprintf("%d", maxWidth)+"s\t%-15s\t%-5s\t%-5s\n", "Function", "Image", "Invocations", "Replicas", "CreatedAt")
		for _, inf := range infs {
			functionImage := inf.Spec.Image
			// if len(function.Image) > 40 {
			// 	functionImage = functionImage[0:38] + ".."
			// }
			cmd.Printf("%-30s\t%-"+fmt.Sprintf("%d", maxWidth)+"s\t%-15d\t%-5d\t\t%-5s\n", inf.Spec.Name, functionImage, int64(inf.Status.InvocationCount), inf.Status.Replicas, inf.Status.CreatedAt.String())
		}
	} else {
		cmd.Printf("%-30s\t%-15s\t%-5s\n", "Function", "Invocations", "Replicas")
		for _, inf := range infs {
			cmd.Printf("%-30s\t%-15d\t%-5d\n", inf.Spec.Name, int64(inf.Status.InvocationCount), inf.Status.Replicas)
		}
	}
	return nil
}

type byName []types.InferenceDeployment

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Spec.Name < a[j].Spec.Name }
