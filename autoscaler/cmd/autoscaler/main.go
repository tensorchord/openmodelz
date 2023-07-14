package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/autoscaler/pkg/autoscalerapp"
	"github.com/tensorchord/openmodelz/autoscaler/pkg/version"
)

func run(args []string) error {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(c.App.Name, version.Package, c.App.Version, version.Revision)
	}

	app := autoscalerapp.New()
	return app.Run(args)
}

func handleErr(err error) {
	if err == nil {
		return
	}

	logrus.Error(err)
}

func main() {
	err := run(os.Args)
	handleErr(err)
}
