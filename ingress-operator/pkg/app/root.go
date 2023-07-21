// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package app

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	controller "github.com/tensorchord/openmodelz/ingress-operator/pkg/controller/v1"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/signals"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/version"
)

const (
	flagDebug = "debug"

	// kubernetes
	flagMasterURL    = "master-url"
	flagKubeConfig   = "kube-config"
	flagQPS          = "kube-qps"
	flagBurst        = "kube-burst"
	flagResyncPeriod = "kube-resync-period"

	// controller
	flagControllerThreads = "controller-thread-count"
	flagNamespace         = "namespace"
	flagHost              = "host"
)

type App struct {
	*cli.App
}

func New() App {
	internalApp := cli.NewApp()
	internalApp.EnableBashCompletion = true
	internalApp.Name = "ingress-operator"
	internalApp.Usage = "kubernetes operator for inference ingress"
	internalApp.HideHelpCommand = true
	internalApp.HideVersion = false
	internalApp.Version = version.GetVersion().String()
	internalApp.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagDebug,
			Usage:   "enable debug output in logs",
			EnvVars: []string{"DEBUG"},
		},
		&cli.StringFlag{
			Name:    flagMasterURL,
			Usage:   "URL to master for kubernetes cluster",
			EnvVars: []string{"MODELZ_MASTER_URL"},
			Aliases: []string{"mu"},
		},
		&cli.StringFlag{
			Name:    flagKubeConfig,
			Usage:   "Path to kubeconfig file. If not provided, will use in-cluster config",
			EnvVars: []string{"MODELZ_KUBE_CONFIG"},
			Aliases: []string{"kc"},
		},
		&cli.IntFlag{
			Name:    flagQPS,
			Usage:   "QPS for kubernetes client",
			Value:   100,
			EnvVars: []string{"MODELZ_KUBE_QPS"},
			Aliases: []string{"kq"},
		},
		&cli.IntFlag{
			Name:    flagBurst,
			Value:   250,
			Usage:   "Burst for kubernetes client",
			EnvVars: []string{"MODELZ_KUBE_BURST"},
			Aliases: []string{"kb"},
		},
		&cli.DurationFlag{
			Name:    flagResyncPeriod,
			Value:   time.Minute * 5,
			Usage:   "Resync period for kubernetes client",
			EnvVars: []string{"MODELZ_KUBE_RESYNC_PERIOD"},
			Aliases: []string{"kr"},
		},
		&cli.IntFlag{
			Name:    flagControllerThreads,
			Value:   1,
			Usage:   "Number of threads to use for controller",
			EnvVars: []string{"MODELZ_CONTROLLER_THREAD_COUNT"},
			Aliases: []string{"ct"},
		},
		&cli.StringFlag{
			Name:    flagNamespace,
			Value:   "default",
			Usage:   "Namespace to create the ingress in. (We need to keep the same namespace as the inference ingress, because kubernetes does not allow cross namespace owner references)",
			EnvVars: []string{"MODELZ_NAMESPACE"},
			Aliases: []string{"ns"},
		},
		&cli.StringFlag{
			Name:    flagHost,
			Value:   "apiserver",
			Usage:   "Host to redirect the request to. (apiserver, agent)",
			EnvVars: []string{"MODELZ_HOST"},
		},
	}
	internalApp.Action = runServer

	// Deal with debug flag.
	var debugEnabled bool

	internalApp.Before = func(context *cli.Context) error {
		debugEnabled = context.Bool(flagDebug)

		if debugEnabled {
			logrus.SetReportCaller(true)
			logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
			logrus.SetLevel(logrus.DebugLevel)
			gin.SetMode(gin.DebugMode)
		} else {
			logrus.SetFormatter(&logrus.JSONFormatter{})
		}

		return nil
	}
	return App{
		App: internalApp,
	}
}

func runServer(clicontext *cli.Context) error {
	c := configFromCLI(clicontext)

	cfgString, _ := c.GetString()
	logrus.WithField("config", c).Info("starting ingress operator")

	if err := c.Validate(); err != nil {
		if clicontext.Bool(flagDebug) {
			return errors.Wrap(err, "invalid config: "+cfgString)
		} else {
			return errors.Wrap(err, "invalid config")
		}
	}

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	s, err := controller.New(c, stopCh)
	if err != nil {
		return errors.Wrap(err, "failed to create server")
	}

	return s.Run(c.Controller.ThreadCount, stopCh)
}
