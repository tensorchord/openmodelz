package server

import (
	"fmt"
	"os/exec"
	"syscall"
)

// k3sInstallStep installs k3s and related tools.
type k3sInstallStep struct {
	options Options
}

func (s *k3sInstallStep) Run() error {
	checkCmd := exec.Command("/bin/sh", "-c", "sudo k3s kubectl get nodes")
	checkCmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	checkCmd.Stdout = nil
	checkCmd.Stderr = nil
	err := checkCmd.Run()
	if err == nil {
		fmt.Fprintf(s.options.OutputStream, "🚧 k3s is already installed, skip...\n")
		return nil
	}

	fmt.Fprintf(s.options.OutputStream, "🚧 Setting up the server...\n")
	// TODO(gaocegege): Embed the script into the binary.
	// Always run start, do not check the hash to decide.
	cmd := exec.Command("/bin/sh", "-c", "curl -sfL https://get.k3s.io | K3S_KUBECONFIG_MODE=644 sh -")
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
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (s *k3sInstallStep) Verify() error {
	return nil
}
