package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

// portForwardCmd represents the port-forward command
var portForwardCmd = &cobra.Command{
	Use:     "port-forward",
	Short:   "Forward one local port to an inference",
	Long:    `Forward one local port to an inference`,
	Example: `  omz port-forward bloomz-560m 7860`,
	GroupID: "debug",
	PreRunE: getAgentClient,
	Args:    cobra.ExactArgs(2),
	RunE:    commandForward,
}

func init() {
	rootCmd.AddCommand(portForwardCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}

func commandForward(cmd *cobra.Command, args []string) error {
	name := args[0]
	port := args[1]

	url, err := url.Parse(fmt.Sprintf("%s/inference/%s.%s", agentURL, name, namespace))
	if err != nil {
		return errors.Newf("failed to parse url: %s", err.Error())
	}
	rp := httputil.NewSingleHostReverseProxy(url)

	cmd.Printf("Forwarding inference %s to local port %s\n", name, port)
	if debug {
		cmd.Println(url)
	}
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			cmd.Printf("Handling connection for %s\n", port)
			p.ServeHTTP(w, r)
		}
	}
	http.HandleFunc("/", handler(rp))
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		return errors.Newf("failed to listen and serve: %s", err.Error())
	}

	return nil
}
