package server

import (
	_ "embed"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"
)

//go:embed registries.yaml
var registriesContent string

const mirrorPath = "/etc/rancher/k3s"
const mirrorFile = "registries.yaml"

// k3sPrepare install everything required by k3s.
type k3sPrepare struct {
	options Options
}

func (s *k3sPrepare) Run() error {
	if !s.options.Mirror.Configured() {
		return nil
	}
	fmt.Fprintf(s.options.OutputStream, "🚧 Configure the mirror...\n")

	tmpl, err := template.New("registries").Parse(registriesContent)
	if err != nil {
		panic(err)
	}
	buf := strings.Builder{}
	err = tmpl.Execute(&buf, s.options.Mirror)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf(
		"sudo mkdir -p %s && sudo tee %s > /dev/null << EOF\n%s\nEOF",
		mirrorPath,
		filepath.Join(mirrorPath, mirrorFile),
		buf.String(),
	))
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
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *k3sPrepare) Verify() error {
	return nil
}
