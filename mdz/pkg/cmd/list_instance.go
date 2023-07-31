package cmd

import (
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

var (
	// Used for flags.
	listInstanceQuiet   bool
	listInstanceVerbose bool
)

// listInstanceCmd represents the list instance command
var listInstanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "List all instances for the given deployment",
	Long:  `List all instances for the given deployment`,
	Example: `  mdz list instance bloomz-560m
  mdz list instance bloomz-560m -v
  mdz list instance bloomz-560m -q`,
	Args:    cobra.ExactArgs(1),
	PreRunE: commandInit,
	RunE:    commandListInstance,
}

func init() {
	listCommand.AddCommand(listInstanceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	listInstanceCmd.Flags().BoolVarP(&listInstanceQuiet, "quiet", "q", false, "Quiet mode - print out only the instance names")
	listInstanceCmd.Flags().BoolVarP(&listInstanceVerbose, "verbose", "v", false, "Verbose mode - print out all instance details")
}

func commandListInstance(cmd *cobra.Command, args []string) error {
	instances, err := agentClient.InstanceList(cmd.Context(), namespace, args[0])
	if err != nil {
		cmd.PrintErrf("Failed to list inference instances: %v\n", err)
		return err
	}

	sort.Sort(byInstanceName(instances))

	if listInstanceQuiet {
		for _, i := range instances {
			cmd.Printf("%s\n", i.Spec.Name)
		}
		return nil
	} else if listInstanceVerbose {
		t := table.NewWriter()
		t.SetStyle(table.Style{
			Box:     table.StyleBoxDefault,
			Color:   table.ColorOptionsDefault,
			Format:  table.FormatOptionsDefault,
			HTML:    table.DefaultHTMLOptions,
			Options: table.OptionsNoBordersAndSeparators,
			Title:   table.TitleOptionsDefault,
		})
		t.AppendHeader(table.Row{"Name", "Status", "Reason", "Message", "CreatedAt"})
		for _, i := range instances {
			t.AppendRow(table.Row{i.Spec.Name, i.Status.Phase,
				i.Status.Reason, i.Status.Message, i.Status.StartTime})
		}
		cmd.Println(t.Render())
		return nil
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
		t.AppendHeader(table.Row{"Name", "Status", "CreatedAt"})
		for _, i := range instances {
			t.AppendRow(table.Row{i.Spec.Name, i.Status.Phase, i.Status.StartTime})
		}
		cmd.Println(t.Render())
		return nil
	}
}

type byInstanceName []types.InferenceDeploymentInstance

func (a byInstanceName) Len() int           { return len(a) }
func (a byInstanceName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byInstanceName) Less(i, j int) bool { return a[i].Spec.Name < a[j].Spec.Name }
