// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package app

import (
	"flag"
	"time"

	"github.com/cockroachdb/errors"
	cli "github.com/urfave/cli/v2"
	"k8s.io/klog"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/controller"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/signals"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/version"
)

const (
	flagDebug = "debug"

	// metrics
	flageMetricsServerPort = "metrics-server-port"

	// kubernetes
	flagMasterURL    = "master-url"
	flagKubeConfig   = "kube-config"
	flagQPS          = "kube-qps"
	flagBurst        = "kube-burst"
	flagResyncPeriod = "kube-resync-period"

	// controller
	flagControllerThreads = "controller-thread-count"

	// huggingface
	flagHuggingfaceEndpoint = "huggingface-endpoint"

	// probes
	flagProbeReadinessInitialDelaySeconds = "probe-readiness-initial-delay-seconds"
	flagProbeReadinessPeriodSeconds       = "probe-readiness-period-seconds"
	flagProbeReadinessTimeoutSeconds      = "probe-readiness-timeout-seconds"

	flagProbeLivenessInitialDelaySeconds = "probe-liveness-initial-delay-seconds"
	flagProbeLivenessPeriodSeconds       = "probe-liveness-period-seconds"
	flagProbeLivenessTimeoutSeconds      = "probe-liveness-timeout-seconds"

	flagProbeStartupInitialDelaySeconds = "probe-startup-initial-delay-seconds"
	flagProbeStartupPeriodSeconds       = "probe-startup-period-seconds"
	flagProbeStartupTimeoutSeconds      = "probe-startup-timeout-seconds"

	// inference
	flagInferenceImagePullPolicy         = "inference-image-pull-policy"
	flagInferenceSetUpRuntimeClassNvidia = "inference-set-up-runtime-class-nvidia"
)

type App struct {
	*cli.App
}

func New() App {
	internalApp := cli.NewApp()
	internalApp.EnableBashCompletion = true
	internalApp.Name = "modelzetes"
	internalApp.Usage = "kubernetes operator for modelz"
	internalApp.HideHelpCommand = true
	internalApp.HideVersion = false
	internalApp.Version = version.GetVersion().String()
	internalApp.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    flagDebug,
			Usage:   "enable debug output in logs",
			EnvVars: []string{"DEBUG"},
		},
		&cli.IntFlag{
			Name:    flageMetricsServerPort,
			Value:   8081,
			Usage:   "port to listen on",
			EnvVars: []string{"MODELZETES_SERVER_PORT"},
			Aliases: []string{"p"},
		},
		&cli.StringFlag{
			Name:    flagMasterURL,
			Usage:   "URL to master for kubernetes cluster",
			EnvVars: []string{"MODELZETES_MASTER_URL"},
			Aliases: []string{"mu"},
		},
		&cli.StringFlag{
			Name:    flagKubeConfig,
			Usage:   "Path to kubeconfig file. If not provided, will use in-cluster config",
			EnvVars: []string{"MODELZETES_KUBE_CONFIG"},
			Aliases: []string{"kc"},
		},
		&cli.IntFlag{
			Name:    flagQPS,
			Usage:   "QPS for kubernetes client",
			Value:   100,
			EnvVars: []string{"MODELZETES_KUBE_QPS"},
			Aliases: []string{"kq"},
		},
		&cli.IntFlag{
			Name:    flagBurst,
			Value:   250,
			Usage:   "Burst for kubernetes client",
			EnvVars: []string{"MODELZETES_KUBE_BURST"},
			Aliases: []string{"kb"},
		},
		&cli.DurationFlag{
			Name:    flagResyncPeriod,
			Value:   time.Minute * 5,
			Usage:   "Resync period for kubernetes client",
			EnvVars: []string{"MODELZETES_KUBE_RESYNC_PERIOD"},
			Aliases: []string{"kr"},
		},
		&cli.IntFlag{
			Name:    flagControllerThreads,
			Value:   1,
			Usage:   "Number of threads to use for controller",
			EnvVars: []string{"MODELZETES_CONTROLLER_THREAD_COUNT"},
			Aliases: []string{"ct"},
		},
		&cli.StringFlag{
			Name: flagHuggingfaceEndpoint,
			Usage: "Endpoint for huggingface modelz API. If not provided, will use " +
				"https://huggingface.co by default",
			EnvVars: []string{"MODELZETES_HUGGINGFACE_ENDPOINT"},
			Aliases: []string{"he"},
		},
		&cli.IntFlag{
			Name:    flagProbeReadinessInitialDelaySeconds,
			Value:   2,
			Usage:   "Initial delay for readiness probe",
			EnvVars: []string{"MODELZETES_PROBE_READINESS_INITIAL_DELAY_SECONDS"},
			Aliases: []string{"prids"},
		},
		&cli.IntFlag{
			Name:    flagProbeReadinessPeriodSeconds,
			Value:   1,
			Usage:   "Period for readiness probe",
			EnvVars: []string{"MODELZETES_PROBE_READINESS_PERIOD_SECONDS"},
			Aliases: []string{"prps"},
		},
		&cli.IntFlag{
			Name:    flagProbeReadinessTimeoutSeconds,
			Value:   1,
			Usage:   "Timeout for readiness probe",
			EnvVars: []string{"MODELZETES_PROBE_READINESS_TIMEOUT_SECONDS"},
			Aliases: []string{"prts"},
		},
		&cli.IntFlag{
			Name:    flagProbeLivenessInitialDelaySeconds,
			Value:   2,
			Usage:   "Initial delay for liveness probe",
			EnvVars: []string{"MODELZETES_PROBE_LIVENESS_INITIAL_DELAY_SECONDS"},
			Aliases: []string{"plids"},
		},
		&cli.IntFlag{
			Name:    flagProbeLivenessPeriodSeconds,
			Value:   1,
			Usage:   "Period for liveness probe",
			EnvVars: []string{"MODELZETES_PROBE_LIVENESS_PERIOD_SECONDS"},
			Aliases: []string{"plps"},
		},
		&cli.IntFlag{
			Name:    flagProbeLivenessTimeoutSeconds,
			Value:   1,
			Usage:   "Timeout for liveness probe",
			EnvVars: []string{"MODELZETES_PROBE_LIVENESS_TIMEOUT_SECONDS"},
			Aliases: []string{"plts"},
		},
		&cli.IntFlag{
			Name:    flagProbeStartupInitialDelaySeconds,
			Value:   0,
			Usage:   "Initial delay for startup probe",
			EnvVars: []string{"MODELZETES_PROBE_STARTUP_INITIAL_DELAY_SECONDS"},
			Aliases: []string{"psids"},
		},
		&cli.IntFlag{
			Name:    flagProbeStartupPeriodSeconds,
			Value:   2,
			Usage:   "Period for startup probe",
			EnvVars: []string{"MODELZETES_PROBE_STARTUP_PERIOD_SECONDS"},
			Aliases: []string{"psps"},
		},
		&cli.IntFlag{
			Name:    flagProbeStartupTimeoutSeconds,
			Value:   1,
			Usage:   "Timeout for startup probe",
			EnvVars: []string{"MODELZETES_PROBE_STARTUP_TIMEOUT_SECONDS"},
			Aliases: []string{"psts"},
		},
		&cli.StringFlag{
			Name:    flagInferenceImagePullPolicy,
			Usage:   "Image pull policy for inference service.",
			Value:   "IfNotPresent",
			EnvVars: []string{"MODELZETES_INFERENCE_IMAGE_PULL_POLICY"},
			Aliases: []string{"iipp"},
		},
		&cli.BoolFlag{
			Name:    flagInferenceSetUpRuntimeClassNvidia,
			Usage:   "If true, will set up the Nvidia RuntimeClassName to the inference deployment.",
			EnvVars: []string{"MODELZETES_INFERENCE_SET_UP_RUNTIME_CLASS_NVIDIA"},
		},
	}
	internalApp.Action = runServer

	// Deal with debug flag.
	var debugEnabled bool

	internalApp.Before = func(context *cli.Context) error {
		debugEnabled = context.Bool(flagDebug)

		fs := flag.NewFlagSet("", flag.PanicOnError)
		klog.InitFlags(fs)

		if debugEnabled {
			fs.Set("v", "10")
		} else {
			fs.Set("v", "0")
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
	klog.V(0).Info("config: ", cfgString)

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
