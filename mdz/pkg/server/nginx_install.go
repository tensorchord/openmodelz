package server

import (
	_ "embed"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

//go:embed nginx-dep.yaml
var nginxYamlContent string

// nginxInstallStep installs the nginx deployment.
type nginxInstallStep struct {
	options Options
}

func (s *nginxInstallStep) Run() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Initializing the load balancer...\n")

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
	if _, err := io.WriteString(stdin, nginxYamlContent); err != nil {
		return err
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *nginxInstallStep) Verify() error {
	fmt.Fprintf(s.options.OutputStream, "🚧 Verifying the load balancer...\n")
	cmd := exec.Command("/bin/sh", "-c", "sudo k3s kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath={@.status.loadBalancer.ingress}")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Debugf("failed to get the ingress ip: %v", err)
		return err
	}
	logrus.Debugf("kubectl get cmd output: %s\n", output)
	if len(output) == 0 {
		return fmt.Errorf("cannot get the ingress ip: output is empty")
	}
	return nil
}
