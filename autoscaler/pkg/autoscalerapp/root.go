// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package autoscalerapp

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/autoscaler/pkg/autoscaler"
	"github.com/tensorchord/openmodelz/autoscaler/pkg/server"
	"github.com/tensorchord/openmodelz/autoscaler/pkg/version"
)

type EnvdServerApp struct {
	*cli.App
}

func New() EnvdServerApp {
	internalApp := cli.NewApp()
	internalApp.EnableBashCompletion = true
	internalApp.Name = "modelz-autoscaler"
	internalApp.Usage = "Autoscaler for modelz serverless inference platform"
	internalApp.HideHelpCommand = true
	internalApp.HideVersion = false
	internalApp.Version = version.GetVersion().String()
	internalApp.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in logs",
		},
		&cli.StringFlag{
			Name:    "gateway-host",
			Usage:   "host for gateway",
			EnvVars: []string{"MODELZ_GATEWAY_HOST"},
			Aliases: []string{"gh"},
		},
		&cli.StringFlag{
			Name:    "prometheus-host",
			Usage:   "host for prometheus",
			Value:   "prometheus",
			EnvVars: []string{"MODELZ_PROMETHEUS_HOST"},
			Aliases: []string{"ph"},
		},
		&cli.IntFlag{
			Name:    "prometheus-port",
			Usage:   "port for prometheus",
			Value:   9090,
			EnvVars: []string{"MODELZ_PROMETHEUS_PORT"},
			Aliases: []string{"pp"},
		},
		&cli.BoolFlag{
			Name:    "basic-auth",
			Usage:   "enable basic auth",
			EnvVars: []string{"MODELZ_BASIC_AUTH"},
			Aliases: []string{"ba"},
			Value:   true,
		},
		&cli.PathFlag{
			Name:    "secret-path",
			Usage:   "path to secrets",
			Value:   "/var/modelz/secrets",
			EnvVars: []string{"MODELZ_SECRET_PATH"},
			Aliases: []string{"sp"},
		},
		&cli.DurationFlag{
			Name:    "interval",
			Usage:   "interval for autoscaling",
			Value:   time.Second,
			EnvVars: []string{"MODELZ_INTERVAL"},
			Aliases: []string{"i"},
		},
	}
	internalApp.Action = runServer

	// Deal with debug flag.
	var debugEnabled bool

	internalApp.Before = func(context *cli.Context) error {
		debugEnabled = context.Bool("debug")

		if debugEnabled {
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
			logrus.SetLevel(logrus.DebugLevel)
		} else {
			logrus.SetFormatter(&logrus.JSONFormatter{})
		}

		return nil
	}
	return EnvdServerApp{
		App: internalApp,
	}
}

func runServer(clicontext *cli.Context) error {
	opt := autoscaler.Opt{
		GatewayHost:      clicontext.String("gateway-host"),
		PrometheusHost:   clicontext.String("prometheus-host"),
		BasicAuthEnabled: clicontext.Bool("basic-auth"),
		SecretPath:       clicontext.Path("secret-path"),
		PrometheusPort:   clicontext.Int("prometheus-port"),
		Interval:         clicontext.Duration("interval"),
	}

	as, err := autoscaler.New(opt)
	if err != nil {
		return errors.Wrap(err, "failed to create autoscaler")
	}

	logrus.Info("starting system info server")
	go server.RunInfoServe()

	logrus.Info("starting autoscaler")
	as.AutoScale(opt.Interval)
	return nil
}
