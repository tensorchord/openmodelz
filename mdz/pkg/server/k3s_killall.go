package server

import (
	"fmt"
	"os/exec"
	"syscall"
)

// k3sKillAllStep installs k3s and related tools.
type k3sKillAllStep struct {
	options Options
}

func (s *k3sKillAllStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Stopping all the processes...\n")
	// TODO(gaocegege): Embed the script into the binary.
	cmd := exec.Command("/bin/sh", "-c", "/usr/local/bin/k3s-killall.sh")
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

func (s *k3sKillAllStep) Verify() error {
	return nil
}
