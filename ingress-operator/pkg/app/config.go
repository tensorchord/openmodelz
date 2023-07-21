package app

import (
	cli "github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/ingress-operator/pkg/config"
)

func configFromCLI(c *cli.Context) config.Config {
	cfg := config.Config{}

	// kubernetes
	cfg.KubeConfig.Kubeconfig = c.String(flagKubeConfig)
	cfg.KubeConfig.MasterURL = c.String(flagMasterURL)
	cfg.KubeConfig.QPS = c.Int(flagQPS)
	cfg.KubeConfig.Burst = c.Int(flagBurst)
	cfg.KubeConfig.ResyncPeriod = c.Duration(flagResyncPeriod)

	// controller
	cfg.Controller.ThreadCount = c.Int(flagControllerThreads)
	cfg.Controller.Namespace = c.String(flagNamespace)
	cfg.Controller.Host = c.String(flagHost)
	return cfg
}
