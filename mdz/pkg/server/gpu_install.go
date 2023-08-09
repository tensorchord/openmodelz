package server

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"syscall"
)

//go:embed gpu-resource.yaml
var gpuYamlContent string

// gpuInstallStep installs the GPU related resources.
type gpuInstallStep struct {
	options Options
}

// check if the Nvidia Toolkit is installed on the host
func (s *gpuInstallStep) hasNvidiaToolkit() bool {
	locations := []string{
		"/usr/local/nvidia/toolkit",
		"/usr/bin",
	}
	binaryNames := []string{
		"nvidia-container-runtime",
		"nvidia-container-runtime-experimental",
	}
	for _, location := range locations {
		for _, name := range binaryNames {
			path := filepath.Join(location, name)
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
	}
	return false
}

func (s *gpuInstallStep) hasNvidiaDevice() bool {
	output, err := exec.Command("/bin/sh", "-c", "lspci").Output()
	if err != nil {
		return false
	}
	regexNvidia := regexp.MustCompile("(?i)nvidia")
	return regexNvidia.Match(output)
}

func (s *gpuInstallStep) Run() error {
	if !s.options.ForceGPU {
		// detect GPU
		if !(s.hasNvidiaDevice() || s.hasNvidiaToolkit()) {
			fmt.Fprintf(s.options.OutputStream, "🚧 Nvidia Toolkit is missing, skip the GPU initialization.\n")
			return nil
		}
	}
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
