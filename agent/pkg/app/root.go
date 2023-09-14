// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package app

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/agent/pkg/server"
	"github.com/tensorchord/openmodelz/agent/pkg/version"
)

const (
	flagDebug = "debug"
	flagDev   = "dev"

	// server
	flagServerPort         = "server-port"
	flagServerReadTimeout  = "server-read-timeout"
	flagServerWriteTimeout = "server-write-timeout"

	// kubernetes
	flagMasterURL    = "master-url"
	flagKubeConfig   = "kube-config"
	flagQPS          = "kube-qps"
	flagBurst        = "kube-burst"
	flagResyncPeriod = "kube-resync-period"

	// inference ingress
	flagIngressEnabled       = "ingress-enabled"
	flagIngressDomain        = "ingress-domain"
	flagIngressNamespace     = "ingress-namespace"
	flagIngressAnyIPToDomain = "ingress-any-ip-to-domain"
	flagIngressTLSEnabled    = "ingress-tls-enabled"

	// inference
	flagInferenceLogTimeout = "inference-log-timeout"
	flagInferenceCacheTTL   = "inference-cache-ttl"

	// build
	flagBuildEnabled         = "build-enabled"
	flagBuilderImage         = "builder-image"
	flagBuildkitdAddress     = "buildkitd-address"
	flagBuildCtlBin          = "buildctl-bin"
	flagBuildRegistry        = "build-registry"
	flagBuildRegistryToken   = "build-registry-token"
	flagBuildImagePullSecret = "build-image-pull-secret"

	// metrics
	flagMetricsPollingInterval = "metrics-polling-interval"
	flagMetricsPort            = "metrics-port"
	flagMetricsPrometheusHost  = "metrics-prometheus-host"
	flagMetricsPrometheusPort  = "metrics-prometheus-port"

	// logs
	flagLogsTimeout   = "logs-timeout"
	flagLogsLokiURL   = "logs-loki-url"
	flaglogsLokiUser  = "logs-loki-user"
	flagLogsLokiToken = "logs-loki-token"

	// db
	flagEventEnabled = "event-enabled"
	flagDBURL        = "db-url"

	// modelz cloud
	flagModelZCloudEnabled                   = "modelz-cloud-enabled"
	flagModelZCloudURL                       = "modelz-cloud-url"
	flagModelZCloudAgentToken                = "modelz-cloud-agent-token"
	flagModelZCloudAgentHeartbeatInterval    = "modelz-cloud-agent-heartbeat-interval"
	flagModelZCloudRegion                    = "modelz-cloud-region"
	flagModelZCloudUnifiedAPIKey             = "modelz-cloud-unified-api-key"
	flagModelZCloudUpstreamTimeout           = "modelz-cloud-upstream-timeout"
	flagModelZCloudMaxIdleConnections        = "modelz-cloud-max-idle-connections"
	flagModelZCloudMaxIdleConnectionsPerHost = "modelz-cloud-max-idle-connections-per-host"
)

type App struct {
	*cli.App
}

func New() App {
	internalApp := cli.NewApp()
	internalApp.EnableBashCompletion = true
	internalApp.Name = "modelz-agent"
	internalApp.Usage = "Cluster agent for modelz"
	internalApp.HideHelpCommand = true
	internalApp.HideVersion = false
	internalApp.Version = version.GetVersion().String()
	internalApp.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  flagDebug,
			Usage: "enable debug output in logs",
		},
		&cli.BoolFlag{
			Name:  flagDev,
			Usage: "enable development mode",
		},
		&cli.IntFlag{
			Name:    flagServerPort,
			Value:   8080,
			Usage:   "port to listen on",
			EnvVars: []string{"MODELZ_AGENT_SERVER_PORT"},
			Aliases: []string{"p"},
		},
		&cli.DurationFlag{
			Name: flagServerReadTimeout,
			Usage: "maximum duration before timing out read of the request, " +
				"including the body",
			Value:   305 * time.Second,
			EnvVars: []string{"MODELZ_AGENT_SERVER_READ_TIMEOUT"},
			Aliases: []string{"srt"},
		},
		&cli.DurationFlag{
			Name: flagServerWriteTimeout,
			Usage: "maximum duration before timing out write of the response, " +
				"including the body",
			Value:   305 * time.Second,
			EnvVars: []string{"MODELZ_AGENT_SERVER_WRITE_TIMEOUT"},
			Aliases: []string{"swt"},
		},
		&cli.StringFlag{
			Name:    flagMasterURL,
			Usage:   "URL to master for kubernetes cluster",
			EnvVars: []string{"MODELZ_AGENT_MASTER_URL"},
			Aliases: []string{"mu"},
		},
		&cli.StringFlag{
			Name:    flagKubeConfig,
			Usage:   "Path to kubeconfig file. If not provided, will use in-cluster config",
			EnvVars: []string{"MODELZ_AGENT_KUBE_CONFIG"},
			Aliases: []string{"kc"},
		},
		&cli.IntFlag{
			Name:    flagQPS,
			Usage:   "QPS for kubernetes client",
			Value:   100,
			EnvVars: []string{"MODELZ_AGENT_KUBE_QPS"},
			Aliases: []string{"kq"},
		},
		&cli.IntFlag{
			Name:    flagBurst,
			Value:   250,
			Usage:   "Burst for kubernetes client",
			EnvVars: []string{"MODELZ_AGENT_KUBE_BURST"},
			Aliases: []string{"kb"},
		},
		&cli.DurationFlag{
			Name:    flagResyncPeriod,
			Value:   time.Hour,
			Usage:   "Resync period for kubernetes client",
			EnvVars: []string{"MODELZ_AGENT_KUBE_RESYNC_PERIOD"},
			Aliases: []string{"kr"},
		},
		&cli.BoolFlag{
			Name: flagIngressEnabled,
			Usage: "Enable inference ingress. " +
				"If enabled, the agent will create ingress for each inference",
			Value:   false,
			EnvVars: []string{"MODELZ_AGENT_INGRESS_ENABLED"},
			Aliases: []string{"ie"},
		},
		&cli.StringFlag{
			Name:    flagIngressDomain,
			Usage:   "Domain for inference ingress",
			Value:   "cloud.modelz.dev",
			EnvVars: []string{"MODELZ_AGENT_INGRESS_DOMAIN"},
			Aliases: []string{"id"},
		},
		&cli.StringFlag{
			Name:    flagIngressNamespace,
			Usage:   "Namespace for inference ingress",
			Value:   "default",
			EnvVars: []string{"MODELZ_AGENT_INGRESS_NAMESPACE"},
			Aliases: []string{"in"},
		},
		&cli.BoolFlag{
			Name: flagIngressAnyIPToDomain,
			Usage: "Enable any ip to domain. " +
				"If enabled, the agent will create ingress for each inference",
			Value:   false,
			EnvVars: []string{"MODELZ_AGENT_INGRESS_ANY_IP_TO_DOMAIN"},
			Aliases: []string{"iad"},
		},
		&cli.BoolFlag{
			Name:    flagIngressTLSEnabled,
			Usage:   "Enable TLS for inference ingress. ",
			Value:   true,
			EnvVars: []string{"MODELZ_AGENT_INGRESS_TLS_ENABLED"},
			Aliases: []string{"it"},
		},
		&cli.DurationFlag{
			Name: flagInferenceLogTimeout,
			Usage: "Timeout for inference log streaming. " +
				"If the inference log has not been updated in this time, " +
				"the connection will be closed.",
			Value:   time.Minute,
			EnvVars: []string{"MODELZ_AGENT_INFERENCE_LOG_TIMEOUT"},
			Aliases: []string{"ilt"},
		},
		&cli.DurationFlag{
			Name:    flagInferenceCacheTTL,
			Usage:   "Time to live for inference cache. ",
			Value:   time.Millisecond * 500,
			EnvVars: []string{"MODELZ_AGENT_INFERENCE_CACHE_TTL"},
			Aliases: []string{"ict"},
		},
		&cli.BoolFlag{
			Name:   flagBuildEnabled,
			Hidden: true,
			Usage: "Enable model build. " +
				"If enabled, the agent will build inference server image",
			Value:   false,
			EnvVars: []string{"MODELZ_AGENT_BUILD_ENABLED"},
			Aliases: []string{"be"},
		},
		&cli.StringFlag{
			Name:   flagBuilderImage,
			Hidden: true,
			Usage: "Image to use for building models. " +
				"Must be a valid docker image reference.",
			EnvVars: []string{"MODELZ_AGENT_BUILDER_IMAGE"},
			Aliases: []string{"bi"},
		},
		&cli.StringFlag{
			Name:   flagBuildkitdAddress,
			Hidden: true,
			Usage: "Address of buildkitd server. " +
				"Must be a valid tcp address.",
			EnvVars: []string{"MODELZ_AGENT_BUILDKITD_ADDRESS"},
			Aliases: []string{"ba"},
		},
		&cli.StringFlag{
			Name:   flagBuildCtlBin,
			Hidden: true,
			Usage: "Path to buildctl binary. " +
				"Must be a valid path to a binary.",
			EnvVars: []string{"MODELZ_AGENT_BUILDCTL_BIN"},
			Aliases: []string{"bb"},
		},
		&cli.StringFlag{
			Name:    flagBuildRegistry,
			Hidden:  true,
			Usage:   "Registry to use for building models. ",
			EnvVars: []string{"MODELZ_AGENT_BUILD_REGISTRY"},
			Aliases: []string{"br"},
		},
		&cli.StringFlag{
			Name:    flagBuildRegistryToken,
			Hidden:  true,
			Usage:   "Token to use for building models. ",
			EnvVars: []string{"MODELZ_AGENT_BUILD_REGISTRY_TOKEN"},
			Aliases: []string{"bt"},
		},
		&cli.StringFlag{
			Name:    flagBuildImagePullSecret,
			Hidden:  true,
			Usage:   "Image pull secret to use for building models.",
			EnvVars: []string{"MODELZ_AGENT_BUILD_IMAGE_PULL_SECRET"},
			Aliases: []string{"bp"},
			Value:   "dockerhub-secret",
		},
		&cli.DurationFlag{
			Name:    flagMetricsPollingInterval,
			Usage:   "Interval to poll metrics from kubernetes",
			Value:   time.Second * 5,
			EnvVars: []string{"MODELZ_AGENT_METRICS_POLLING_INTERVAL"},
			Aliases: []string{"mpi"},
		},
		&cli.IntFlag{
			Name:    flagMetricsPort,
			Usage:   "Port to expose metrics on. ",
			Value:   8082,
			EnvVars: []string{"MODELZ_AGENT_METRICS_PORT"},
			Aliases: []string{"mp"},
		},
		&cli.StringFlag{
			Name:    flagMetricsPrometheusHost,
			Value:   "localhost",
			Usage:   "Host to expose prometheus metrics on. ",
			EnvVars: []string{"MODELZ_AGENT_METRICS_PROMETHEUS_HOST"},
			Aliases: []string{"mph"},
		},
		&cli.IntFlag{
			Name:    flagMetricsPrometheusPort,
			Usage:   "Port to expose prometheus metrics on. ",
			Value:   9090,
			EnvVars: []string{"MODELZ_AGENT_METRICS_PROMETHEUS_PORT"},
			Aliases: []string{"mpp"},
		},
		&cli.DurationFlag{
			Name:    flagLogsTimeout,
			Usage:   "request timeout to query the logs",
			Value:   time.Second * 5,
			EnvVars: []string{"MODELZ_AGENT_LOGS_TIMEOUT"},
		},
		&cli.StringFlag{
			Name:    flagLogsLokiURL,
			Hidden:  true,
			Usage:   "Loki service URL",
			EnvVars: []string{"MODELZ_AGENT_LOGS_LOKI_URL"},
		},
		&cli.StringFlag{
			Name:    flaglogsLokiUser,
			Hidden:  true,
			Usage:   "Loki service auth user",
			EnvVars: []string{"MODELZ_AGENT_LOGS_LOKI_USER"},
		},
		&cli.StringFlag{
			Name:    flagLogsLokiToken,
			Hidden:  true,
			Usage:   "Loki service auth token",
			EnvVars: []string{"MODELZ_AGENT_LOGS_LOKI_TOKEN"},
		},
		&cli.BoolFlag{
			Name:    flagEventEnabled,
			Hidden:  true,
			Usage:   "Enable event logging",
			Value:   false,
			EnvVars: []string{"MODELZ_AGENT_EVENT_ENABLED"},
			Aliases: []string{"ee"},
		},
		&cli.StringFlag{
			Name:    flagDBURL,
			Usage:   "Postgres database URL",
			Hidden:  true,
			Aliases: []string{"du"},
			EnvVars: []string{"MODELZ_AGENT_DB_URL"},
		},
		&cli.BoolFlag{
			Name:    flagModelZCloudEnabled,
			Usage:   "Enable modelz cloud, agent as modelz cloud agent",
			Value:   false,
			EnvVars: []string{"MODELZ_AGENT_MODELZ_CLOUD_ENABLED"},
			Aliases: []string{"mzc"},
		},
		&cli.StringFlag{
			Name:    flagModelZCloudURL,
			Usage:   "Modelz cloud URL",
			EnvVars: []string{"MODELZ_AGENT_MODELZ_CLOUD_URL"},
			Aliases: []string{"mzu"},
			Value:   "https://cloud.modelz.ai",
		},
		&cli.StringFlag{
			Name:    flagModelZCloudAgentToken,
			Usage:   "Modelz cloud agent token",
			EnvVars: []string{"MODELZ_CLOUD_AGENT_TOKEN"},
			Aliases: []string{"mzt"},
		},
		&cli.DurationFlag{
			Name:    flagModelZCloudAgentHeartbeatInterval,
			Usage:   "Modelz cloud agent heartbeat interval",
			EnvVars: []string{"MODELZ_CLOUD_AGENT_HEARTBEAT_INTERVAL"},
			Aliases: []string{"mzh"},
			Value:   time.Minute * 1,
		},
		&cli.StringFlag{
			Name:    flagModelZCloudRegion,
			Usage:   "Modelz cloud agent region",
			EnvVars: []string{"MODELZ_CLOUD_AGENT_REGION"},
			Aliases: []string{"mzr"},
			Value:   "us-central1",
		},
		&cli.StringFlag{
			Name:    flagModelZCloudUnifiedAPIKey,
			Usage:   "Modelz cloud agent unified api key",
			EnvVars: []string{"MODELZ_CLOUD_AGENT_UNIFIED_API_KEY"},
			Aliases: []string{"mzua"},
		},
		&cli.DurationFlag{
			Name:    flagModelZCloudUpstreamTimeout,
			Usage:   "upstream timeout",
			EnvVars: []string{"MODELZ_UPSTREAM_TIMEOUT"},
			Aliases: []string{"ut"},
			Value:   300 * time.Second,
		},
		&cli.IntFlag{
			Name:    flagModelZCloudMaxIdleConnections,
			Usage:   "max idle connections",
			EnvVars: []string{"MODELZ_MAX_IDLE_CONNECTIONS"},
			Aliases: []string{"mic"},
			Value:   1024,
		},
		&cli.IntFlag{
			Name:    flagModelZCloudMaxIdleConnectionsPerHost,
			Usage:   "max idle connections per host",
			EnvVars: []string{"MODELZ_MAX_IDLE_CONNECTIONS_PER_HOST"},
			Aliases: []string{"mich"},
			Value:   1024,
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

	if clicontext.Bool(flagDebug) {
		logrus.Debug("debug mode enabled")
		cfgString, _ := c.GetString()
		logrus.WithField("config", cfgString).Debug("config")
	}

	if err := c.Validate(); err != nil {
		if clicontext.Bool(flagDebug) {
			logrus.WithError(err).Error("invalid config")
		} else {
			return errors.Wrap(err, "invalid config")
		}
	}

	s, err := server.New(c)
	if err != nil {
		return errors.Wrap(err, "failed to create server")
	}

	return s.Run()
}
