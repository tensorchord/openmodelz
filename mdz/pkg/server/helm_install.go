package server

import (
	_ "embed"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

//go:embed openmodelz.yaml
var yamlContent string

// helmStep installs the OpenModelZ deployments using Helm CRD.
type helmStep struct {
	options Options
}

func (s *helmStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Initializing the server...\n")
	cmd := exec.Command("/bin/sh", "-c", "sudo k3s kubectl apply -f -")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close() // the doc says subProcess.Wait will close it, but I'm not sure, so I kept this line
	if s.options.Verbose {
		cmd.Stderr = s.options.OutputStream
		cmd.Stdout = s.options.OutputStream
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := io.WriteString(stdin, yamlContent); err != nil {
		return err
	}
	stdin.Close()

	fmt.Fprintf(s.options.OutputStream, "🚧 Waiting for the server to be ready...\n")
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *helmStep) Verify() error {
	return nil
}
