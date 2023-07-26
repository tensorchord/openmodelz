package cmd

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// portForwardCmd represents the port-forward command
var portForwardCmd = &cobra.Command{
	Use:     "port-forward",
	Short:   "Forward one local port to a deployment",
	Long:    `Forward one local port to a deployment`,
	Example: `  mdz port-forward blomdz-560m 7860`,
	GroupID: "debug",
	PreRunE: commandInit,
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

	if _, err := agentClient.InferenceGet(cmd.Context(), namespace, name); err != nil {
		cmd.PrintErrf("Failed to get inference: %s\n", errors.Cause(err))
		return err
	}

	url, err := url.Parse(fmt.Sprintf("%s/inference/%s.%s", mdzURL, name, namespace))
	if err != nil {
		cmd.PrintErrf("Failed to parse URL: %s\n", errors.Cause(err))
		return errors.Newf("failed to parse URL: %s\n", errors.Cause(err))
	}
	rp := httputil.NewSingleHostReverseProxy(url)

	cmd.Printf("Forwarding inference %s to local port %s\n", name, port)
	logrus.WithField("url", url).Debugf(
		"Forwarding inference %s to local port %s\n", name, port)
	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			cmd.Printf("Handling connection for %s\n", port)
			p.ServeHTTP(w, r)
		}
	}
	http.HandleFunc("/", handler(rp))
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		cmd.PrintErrf("Failed to listen and serve: %s\n", errors.Cause(err))
		return errors.Newf("failed to listen and serve: %s", errors.Cause(err))
	}

	return nil
}
