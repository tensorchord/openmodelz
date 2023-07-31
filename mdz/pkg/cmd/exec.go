package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tensorchord/openmodelz/agent/client"
	"github.com/tensorchord/openmodelz/mdz/pkg/cmd/streams"
	terminal "golang.org/x/term"
	"k8s.io/apimachinery/pkg/util/rand"
)

var (
	execInstance    string
	execTTY         bool
	execInteractive bool
)

// execCommand represents the exec command
var execCommand = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command in a deployment",
	Long:  `Execute a command in a deployment. If no instance is specified, the first instance is used.`,
	Example: `  mdz exec bloomz-560m ps
  mdz exec bloomz-560m --instance bloomz-560m-abcde-abcde ps
  mdz exec bllomz-560m -ti bash
  mdz exec bloomz-560m --instance bloomz-560m-abcde-abcde -ti bash`,
	GroupID: "debug",
	PreRunE: commandInit,
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
	execCommand.Flags().StringVarP(&execInstance, "instance", "s", "", "Instance name")
	execCommand.Flags().BoolVarP(&execTTY, "tty", "t", false, "Allocate a TTY for the container")
	execCommand.Flags().BoolVarP(&execInteractive, "interactive", "i", false, "Keep stdin open even if not attached")
}

func commandExec(cmd *cobra.Command, args []string) error {
	name := args[0]

	if execInstance == "" {
		instances, err := agentClient.InstanceList(cmd.Context(), namespace, name)
		if err != nil {
			cmd.PrintErrf("Failed to list instances: %s\n", errors.Cause(err))
			return err
		}
		if len(instances) == 0 {
			cmd.PrintErrf("instance %s not found\n", name)
			return errors.Newf("instance %s not found", name)
		} else if len(instances) > 1 {
			cmd.PrintErrf("inference %s has multiple instances, please specify with -i\n", name)
			return errors.Newf("inference %s has multiple instances, please specify with -i", name)
		}
		execInstance = instances[0].Spec.Name
	}

	if execTTY {
		shell := "sh"
		if len(args) > 1 {
			shell = args[1]
		} else if len(args) > 2 {
			cmd.PrintErrf("too many arguments in tty mode, please use a shell program e.g. bash\n")
			return fmt.Errorf("too many arguments")
		}

		if !isAvailableShell(shell) {
			cmd.PrintErrf("shell %s is not available, try `sh` or `bash`\n", shell)
			return fmt.Errorf("shell %s is not available, try `sh` or `bash`", shell)
		}

		resp, err := agentClient.InstanceExecTTY(cmd.Context(), namespace, name, execInstance, []string{shell})
		if err != nil {
			cmd.PrintErrf("Failed to execute the shell: %s\n", errors.Cause(err))
			return err
		}
		defer resp.Conn.Close()
		c := resp.Conn

		if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
			cmd.PrintErrf("stdin/stdout should be terminal\n")
			return fmt.Errorf("stdin/stdout should be terminal")
		}
		// oldState, err := terminal.MakeRaw(0)
		// if err != nil {
		// 	cmd.PrintErrf("Failed to make raw terminal: %s\n", errors.Cause(err))
		// 	return err
		// }
		// oldOutState, err := terminal.MakeRaw(1)
		// if err != nil {
		// 	cmd.PrintErrf("Failed to make raw terminal: %s\n", errors.Cause(err))
		// 	return err
		// }
		// defer func() {
		// 	terminal.Restore(0, oldState)
		// 	terminal.Restore(1, oldOutState)
		// }()

		// Send terminal size.
		w, h, err := terminal.GetSize(0)
		if err != nil {
			cmd.PrintErrf("Failed to get terminal size: %s\n", errors.Cause(err))
			return err
		}
		msg := &client.TerminalMessage{
			ID:   rand.String(5),
			Op:   "resize",
			Data: "",
			Rows: uint16(h),
			Cols: uint16(w),
		}
		if err := c.WriteJSON(msg); err != nil {
			cmd.PrintErrf("Failed to send terminal message: %s\n", errors.Cause(err))
			return err
		}

		errCh := make(chan error, 1)
		cli := newMDZCLI()

		go func() {
			defer close(errCh)
			errCh <- func() error {
				streamer := hijackedIOStreamer{
					streams:      cli,
					inputStream:  cli.In(),
					outputStream: cli.Out(),
					errorStream:  cli.Err(),
					resp:         resp,
					tty:          true,
					detachKeys:   "",
				}

				return streamer.stream(cmd.Context())
			}()
		}()

		if err := <-errCh; err != nil {
			logrus.Debugf("Error hijack: %s", err)
			return err
		}

		return nil
	} else {
		res, err := agentClient.InstanceExec(cmd.Context(), namespace, name, execInstance, args[1:], false)
		if err != nil {
			cmd.PrintErrf("Failed to execute the command: %s\n", errors.Cause(err))
			return err
		}

		cmd.Printf("%s", res)
		return nil
	}
}

func isAvailableShell(shell string) bool {
	switch shell {
	case "sh", "bash", "zsh", "fish":
		return true
	default:
		return false
	}
}

type mdzCli struct {
	in  *streams.In
	out *streams.Out
	err io.Writer
}

func newMDZCLI() *mdzCli {
	return &mdzCli{
		in:  streams.NewIn(os.Stdin),
		out: streams.NewOut(os.Stdout),
		err: os.Stderr,
	}
}

func (c mdzCli) In() *streams.In {
	return c.in
}

func (c mdzCli) Out() *streams.Out {
	return c.out
}

func (c mdzCli) Err() io.Writer {
	return c.err
}
