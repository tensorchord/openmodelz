package server

import (
	_ "embed"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

//go:embed k3s-install.sh
var bashContent string

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
		fmt.Fprintf(s.options.OutputStream, "🚧 The server is already created, skip...\n")
		return nil
	}

	fmt.Fprintf(s.options.OutputStream, "🚧 Creating the server...\n")
	// TODO(gaocegege): Embed the script into the binary.
	// Always run start, do not check the hash to decide.
	cmd := exec.Command("/bin/sh", "-c", "INSTALL_K3S_VERSION=v1.27.3+k3s1 INSTALL_K3S_EXEC='--disable=traefik' INSTALL_K3S_FORCE_RESTART=true K3S_KUBECONFIG_MODE=644 K3S_TOKEN=openmodelz sh -")
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

	if _, err := io.WriteString(stdin, bashContent); err != nil {
		return err
	}
	// Close the input stream to finish the pipe. Then the command will use the
	// input from the pipe to start the next process.
	stdin.Close()

	fmt.Fprintf(s.options.OutputStream, "🚧 Waiting for the server to be created...\n")
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *k3sInstallStep) Verify() error {
	return nil
}
