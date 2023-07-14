package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"
	"k8s.io/klog"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/app"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/version"
)

func run(args []string) error {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(c.App.Name, version.Package, c.App.Version, version.Revision)
	}
	klog.InitFlags(nil)

	a := app.New()
	return a.Run(args)
}

func handleErr(err error) {
	if err == nil {
		return
	}

	klog.Error(err)
	os.Exit(1)
}

func main() {
	err := run(os.Args)
	handleErr(err)
}
