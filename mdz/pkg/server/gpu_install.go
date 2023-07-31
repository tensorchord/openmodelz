package server

import (
	_ "embed"
	"fmt"
	"io"
	"os/exec"
	"syscall"
)

//go:embed gpu-resource.yaml
var gpuYamlContent string

// gpuInstallStep installs the GPU related resources.
type gpuInstallStep struct {
	options Options
}

func (s *gpuInstallStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Initializing the GPU resource...\n")

	cmd := exec.Command("/bin/sh", "-c", "sudo k3s kubectl apply -f -")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
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
	if _, err := io.WriteString(stdin, gpuYamlContent); err != nil {
		return err
	}
	// Close the input stream to finish the pipe. Then the command will use the
	// input from the pipe to start the next process.
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *gpuInstallStep) Verify() error {
	return nil
}
