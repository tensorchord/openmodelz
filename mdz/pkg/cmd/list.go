package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/mdz/pkg/telemetry"
)

const (
	annotationDomain = "ai.tensorchord.domain"
)

var (
	// Used for flags.
	listQuiet   bool
	listVerbose bool
)

// listCommand represents the list command
var listCommand = &cobra.Command{
	Use:   "list",
	Short: "List the deployments",
	Long:  `List the deployments`,
	Example: `  mdz list
  mdz list -v
  mdz list -q`,
	GroupID: "basic",
	PreRunE: commandInit,
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
	telemetry.GetTelemetry().Record("list")
	infs, err := agentClient.InferenceList(cmd.Context(), namespace)
	if err != nil {
		cmd.PrintErrf("Failed to list inferences: %v\n", err)
		return err
	}

	sort.Sort(byName(infs))
	if listQuiet {
		for _, inf := range infs {
			cmd.Printf("%s\n", inf.Spec.Name)
		}
		return nil
	} else if listVerbose {
		t := table.NewWriter()
		t.SetStyle(table.Style{
			Box:     table.StyleBoxDefault,
			Color:   table.ColorOptionsDefault,
			Format:  table.FormatOptionsDefault,
			HTML:    table.DefaultHTMLOptions,
			Options: table.OptionsNoBordersAndSeparators,
			Title:   table.TitleOptionsDefault,
		})
		t.AppendHeader(table.Row{"Name", "Endpoint", "Image", "Status", "Invocations", "Replicas", "CreatedAt"})

		for _, inf := range infs {
			functionImage := inf.Spec.Image
			createdAt := ""
			if inf.Status.CreatedAt != nil {
				createdAt = inf.Status.CreatedAt.String()
			}
			t.AppendRow(table.Row{
				inf.Spec.Name,
				getEndpoint(inf),
				functionImage,
				inf.Status.Phase,
				int64(inf.Status.InvocationCount),
				fmt.Sprintf("%d/%d", inf.Status.AvailableReplicas, inf.Status.Replicas),
				createdAt,
			})
		}

		cmd.Println(t.Render())
	} else {
		t := table.NewWriter()
		t.SetStyle(table.Style{
			Box:     table.StyleBoxDefault,
			Color:   table.ColorOptionsDefault,
			Format:  table.FormatOptionsDefault,
			HTML:    table.DefaultHTMLOptions,
			Options: table.OptionsNoBordersAndSeparators,
			Title:   table.TitleOptionsDefault,
		})
		t.AppendHeader(table.Row{"Name", "Endpoint", "Status", "Invocations", "Replicas"})
		for _, inf := range infs {
			t.AppendRow(table.Row{
				inf.Spec.Name,
				getEndpoint(inf),
				inf.Status.Phase,
				int64(inf.Status.InvocationCount),
				fmt.Sprintf("%d/%d", inf.Status.AvailableReplicas, inf.Status.Replicas),
			})
		}
		cmd.Println(t.Render())
	}
	return nil
}

type byName []types.InferenceDeployment

func (a byName) Len() int           { return len(a) }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byName) Less(i, j int) bool { return a[i].Spec.Name < a[j].Spec.Name }

func getEndpoint(inf types.InferenceDeployment) string {
	endpoint := fmt.Sprintf("%s/inference/%s.%s", mdzURL, inf.Spec.Name, inf.Spec.Namespace)
	if d, ok := inf.Spec.Annotations[annotationDomain]; ok {
		// Replace https with http now.
		rawHTTPDomain := strings.Replace(d, "https://", "http://", 1)
		endpoint = fmt.Sprintf("%s\n%s", rawHTTPDomain, endpoint)
	}
	return endpoint
}
