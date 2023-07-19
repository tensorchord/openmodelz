package cmd

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/apimachinery/pkg/util/rand"
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
		u := url.URL{Scheme: "ws", Host: "localhost:8081", Path: "/system/inference/llm/instance/llm-8598f68565-45zmn/exec", RawQuery: "namespace=default&tty=true&command=bash"}
		c, _, err := websocket.DefaultDialer.DialContext(cmd.Context(), u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close()

		cmd.Printf("Connected to %s\n", execInstance)

		if !terminal.IsTerminal(0) || !terminal.IsTerminal(1) {
			return fmt.Errorf("stdin/stdout should be terminal")
		}
		oldState, err := terminal.MakeRaw(0)
		if err != nil {
			return err
		}
		defer terminal.Restore(0, oldState)

		screen := struct {
			io.Reader
			io.Writer
		}{os.Stdin, os.Stdout}
		term := terminal.NewTerminal(screen, "")
		go func() {
			for {
				line, err := term.ReadLine()
				if err != nil {
					panic(err)
				}
				if line == "" {
					continue
				}
				msg := &TerminalMessage{
					ID:   rand.String(5),
					Op:   "stdin",
					Data: line + "\n",
				}
				if err := c.WriteJSON(msg); err != nil {
					panic(err)
				}
			}
		}()

		for {
			var msg TerminalMessage
			if err := c.ReadJSON(&msg); err != nil {
				return err
			}
			cmd.Printf("%s", msg.Data)
		}
	} else {
		res, err := agentClient.InstanceExec(cmd.Context(), namespace, name, execInstance, args[1:], false)
		if err != nil {
			return err
		}

		cmd.Printf("%s", res)
		return nil
	}
}

// TerminalMessage is the messaging protocol between ShellController and TerminalSession.
//
// OP      DIRECTION  FIELD(S) USED  DESCRIPTION
// ---------------------------------------------------------------------
// bind    fe->be     SessionID      Id sent back from TerminalResponse
// stdin   fe->be     Data           Keystrokes/paste buffer
// resize  fe->be     Rows, Cols     New terminal size
// stdout  be->fe     Data           Output from the process
// toast   be->fe     Data           OOB message to be shown to the user
type TerminalMessage struct {
	ID   string `json:"id,omitempty"`
	Op   string `json:"op,omitempty"`
	Data string `json:"data,omitempty"`
	Rows uint16 `json:"rows,omitempty"`
	Cols uint16 `json:"cols,omitempty"`
}
