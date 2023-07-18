package cmd

import (
	"bufio"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var (
	execInstance string
	execTTY      bool
)

// execCommand represents the exec command
var execCommand = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command in an inference",
	Long:  `Execute a command in an inference`,
	Example: `  mdz exec bloomz-560m ps
  mdz exec bloomz-560m -i bloomz-560m-abcde-abcde ps`,
	GroupID: "debug",
	PreRunE: getAgentClient,
	Args:    cobra.MinimumNArgs(1),
	RunE:    commandExec,
}

func init() {
	rootCmd.AddCommand(execCommand)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	execCommand.Flags().StringVarP(&execInstance, "instance", "i", "", "Instance name")
	execCommand.Flags().BoolVarP(&execTTY, "tty", "t", false, "Allocate a TTY for the container")
}

func commandExec(cmd *cobra.Command, args []string) error {
	name := args[0]

	if execInstance == "" {
		instances, err := agentClient.InstanceList(cmd.Context(), namespace, name)
		if err != nil {
			return err
		}
		if len(instances) == 0 {
			return errors.Newf("instance %s not found", name)
		} else if len(instances) > 1 {
			return errors.Newf("inference %s has multiple instances, please specify with -i", name)
		}
		execInstance = instances[0].Spec.Name
	}

	if execTTY {
		resp, err := agentClient.InstanceExecTTY(cmd.Context(), namespace, name, execInstance, args[1:])
		if err != nil {
			return err
		}
		defer resp.Close()
		resp.Conn.Write([]byte("ls\r"))
		scanner := bufio.NewScanner(resp.Conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		return nil
	} else {
		res, err := agentClient.InstanceExec(cmd.Context(), namespace, name, execInstance, args[1:], false)
		if err != nil {
			return err
		}

		cmd.Printf("%s", res)
		return nil
	}
}
