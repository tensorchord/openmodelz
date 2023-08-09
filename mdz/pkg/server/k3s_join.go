package server

import (
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

// k3sJoinStep installs k3s and related tools.
type k3sJoinStep struct {
	options Options
}

func (s *k3sJoinStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Joining the cluster...\n")
	// TODO(gaocegege): Embed the script into the binary.
	cmdStr := fmt.Sprintf("INSTALL_K3S_FORCE_RESTART=true K3S_KUBECONFIG_MODE=644 K3S_TOKEN=openmodelz K3S_URL=https://%s:6443 sh -", s.options.ServerIP)

	cmd := exec.Command("/bin/sh", "-c", cmdStr)
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

	fmt.Fprintf(s.options.OutputStream, "🚧 Waiting for the server to be ready...\n")
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *k3sJoinStep) Verify() error {
	return nil
}
