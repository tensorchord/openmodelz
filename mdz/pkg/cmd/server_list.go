package cmd

import (
	"fmt"
	"math"

	"github.com/cockroachdb/errors"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/mdz/pkg/telemetry"
)

var (
	// Used for flags.
	serverListQuiet   bool
	serverListVerbose bool
)

// serverListCmd represents the server list command
var serverListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all servers in the cluster",
	Long:    `List all servers in the cluster`,
	Example: `  mdz server list`,
	PreRunE: commandInit,
	RunE:    commandServerList,
}

func init() {
	serverCmd.AddCommand(serverListCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	serverListCmd.Flags().BoolVarP(&serverListQuiet, "quiet", "q", false, "Quiet mode - print out only the server names")
	serverListCmd.Flags().BoolVarP(&serverListVerbose, "verbose", "v", false, "Verbose mode - print out all server details")
}

func commandServerList(cmd *cobra.Command, args []string) error {
	telemetry.GetTelemetry().Record("server list")
	servers, err := agentClient.ServerList(cmd.Context())
	if err != nil {
		cmd.PrintErrf("Failed to list servers: %s\n", errors.Cause(err))
		return err
	}

	if serverListQuiet {
		for _, server := range servers {
			cmd.Printf("%s\n", server.Spec.Name)
		}
	} else if serverListVerbose {
		t := table.NewWriter()
		t.SetStyle(table.Style{
			Box:     table.StyleBoxDefault,
			Color:   table.ColorOptionsDefault,
			Format:  table.FormatOptionsDefault,
			HTML:    table.DefaultHTMLOptions,
			Options: table.OptionsNoBordersAndSeparators,
			Title:   table.TitleOptionsDefault,
		})
		t.AppendHeader(table.Row{"Name", "Phase", "Allocatable", "Capacity", "Distribution", "OS", "Kernel", "Labels"})

		for _, server := range servers {
			t.AppendRow(table.Row{server.Spec.Name, server.Status.Phase,
				resourceListString(server.Status.Allocatable),
				resourceListString(server.Status.Capacity),
				server.Status.System.OSImage,
				server.Status.System.OperatingSystem,
				server.Status.System.KernelVersion,
				labelsString(server.Spec.Labels),
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
		t.AppendHeader(table.Row{"Name", "Phase", "Allocatable", "Capacity"})

		for _, server := range servers {
			t.AppendRow(table.Row{server.Spec.Name, server.Status.Phase,
				resourceListString(server.Status.Allocatable),
				resourceListString(server.Status.Capacity)})
		}
		cmd.Println(t.Render())
	}
	return nil
}

func labelsString(labels map[string]string) string {
	res := ""
	for k, v := range labels {
		res += fmt.Sprintf("%s=%s\n", k, v)
	}
	if len(res) == 0 {
		return res
	}
	return res[:len(res)-1]
}

func prettyByteSize(quantity string) (string, error) {
	r, err := resource.ParseQuantity(quantity)
	if err != nil {
		return "", err
	}
	bf := float64(r.Value())
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit), nil
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fPiB", bf), nil
}

func resourceListString(l types.ResourceList) string {
	res := fmt.Sprintf("cpu: %s", l[types.ResourceCPU])
	memory, ok := l[types.ResourceMemory]
	if ok {
		prettyMem, err := prettyByteSize(string(memory))
		if err != nil {
			logrus.Infof("failed to parse the memory quantity: %s", memory)
		} else {
			memory = types.Quantity(prettyMem)
		}
	}
	res += fmt.Sprintf("\nmemory: %s", memory)
	if l[types.ResourceGPU] != "" {
		res += fmt.Sprintf("\ngpu: %s", l[types.ResourceGPU])
	}
	return res
}
