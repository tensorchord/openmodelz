package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/agent/pkg/app"
	"github.com/tensorchord/openmodelz/agent/pkg/version"
)

func run(args []string) error {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Println(c.App.Name, version.Package, c.App.Version, version.Revision)
	}

	app := app.New()
	return app.Run(args)
}

func handleErr(err error) {
	if err == nil {
		return
	}

	logrus.Error(err)
	os.Exit(1)
}

// @title       modelz cluster agent
// @version     v0.0.23
// @description modelz kubernetes cluster agent

// @contact.name  modelz support
// @contact.url   https://github.com/tensorchord/openmodelz
// @contact.email modelz-support@tensorchord.ai

// @host     localhost:8081
// @BasePath /
// @schemes  http
func main() {
	err := run(os.Args)
	handleErr(err)
}
