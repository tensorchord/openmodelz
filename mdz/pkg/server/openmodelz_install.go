package server

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

//go:embed openmodelz.yaml
var yamlContent string

// openModelZInstallStep installs the OpenModelZ deployments.
type openModelZInstallStep struct {
	options Options
}

func (s *openModelZInstallStep) Run() error {
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

	variables := struct {
		Domain     string
		IpToDomain bool
	}{}
	if s.options.Domain != nil {
		variables.Domain = *s.options.Domain
		variables.IpToDomain = false
	} else {
		fmt.Fprintf(s.options.OutputStream, "🚧 No domain provided, using the server IP...\n")
		variables.Domain = ""
		variables.IpToDomain = true
	}
	tmpl, err := template.New("openmodelz").Parse(yamlContent)
	if err != nil {
		panic(err)
	}
	buf := strings.Builder{}
	err = tmpl.Execute(&buf, variables)
	if err != nil {
		panic(err)
	}

	if _, err := io.WriteString(stdin, buf.String()); err != nil {
		return err
	}
	stdin.Close()

	fmt.Fprintf(s.options.OutputStream, "🚧 Waiting for the server to be ready...\n")
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

func (s *openModelZInstallStep) Verify() error {
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
