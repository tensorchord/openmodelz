package server

import (
	"fmt"
	"os/exec"
	"syscall"
)

type agentDRunStep struct {
	options Options
}

// TODO(gaocegege): There is still a bug, thus it cannot be used actually.
// The process will exit after the command returns. We need to put it in systemd.
func (s *agentDRunStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Running the agent for docker runtime...\n")
	cmd := exec.Command("/bin/sh", "-c", "mdz local-agent &")
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

func (s *agentDRunStep) Verify() error {
	return nil
}
