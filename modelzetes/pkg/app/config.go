package app

import (
	cli "github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/config"
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

	// huggingface
	cfg.HuggingfaceProxy.Endpoint = c.String(flagHuggingfaceEndpoint)

	// probes
	cfg.Probes.Readiness.InitialDelaySeconds = c.Int(flagProbeReadinessInitialDelaySeconds)
	cfg.Probes.Readiness.PeriodSeconds = c.Int(flagProbeReadinessPeriodSeconds)
	cfg.Probes.Readiness.TimeoutSeconds = c.Int(flagProbeReadinessTimeoutSeconds)

	cfg.Probes.Liveness.InitialDelaySeconds = c.Int(flagProbeLivenessInitialDelaySeconds)
	cfg.Probes.Liveness.PeriodSeconds = c.Int(flagProbeLivenessPeriodSeconds)
	cfg.Probes.Liveness.TimeoutSeconds = c.Int(flagProbeLivenessTimeoutSeconds)

	cfg.Probes.Startup.InitialDelaySeconds = c.Int(flagProbeStartupInitialDelaySeconds)
	cfg.Probes.Startup.PeriodSeconds = c.Int(flagProbeStartupPeriodSeconds)
	cfg.Probes.Startup.TimeoutSeconds = c.Int(flagProbeStartupTimeoutSeconds)

	// inference
	cfg.Inference.ImagePullPolicy = c.String(flagInferenceImagePullPolicy)
	return cfg
}
