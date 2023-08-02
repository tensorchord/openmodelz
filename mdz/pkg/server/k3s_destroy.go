package server

import (
	"fmt"
	"os/exec"
	"syscall"
)

// k3sDestroyAllStep installs k3s and related tools.
type k3sDestroyAllStep struct {
	options Options
}

func (s *k3sDestroyAllStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Destroy the OpenModelz Cluster...\n")
	// TODO(gaocegege): Embed the script into the binary.
	cmd := exec.Command("/bin/sh", "-c", "/usr/local/bin/k3s-uninstall.sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	if s.options.Verbose {
		cmd.Stderr = s.options.OutputStream
		cmd.Stdout = s.options.OutputStream
	} else {
		cmd.Stdout = nil
		cmd.Stderr = nil
	}
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *k3sDestroyAllStep) Verify() error {
	return nil
}
