package app

import (
	"github.com/urfave/cli/v2"

	"github.com/tensorchord/openmodelz/agent/pkg/config"
)

func configFromCLI(c *cli.Context) config.Config {
	cfg := config.New()

	// server
	cfg.Server.Dev = c.Bool(flagDev)
	cfg.Server.ServerPort = c.Int(flagServerPort)
	cfg.Server.ReadTimeout = c.Duration(flagServerReadTimeout)
	cfg.Server.WriteTimeout = c.Duration(flagServerWriteTimeout)

	// kubernetes
	cfg.KubeConfig.Kubeconfig = c.String(flagKubeConfig)
	cfg.KubeConfig.MasterURL = c.String(flagMasterURL)
	cfg.KubeConfig.QPS = c.Int(flagQPS)
	cfg.KubeConfig.Burst = c.Int(flagBurst)
	cfg.KubeConfig.ResyncPeriod = c.Duration(flagResyncPeriod)

	// inference ingress
	cfg.Ingress.IngressEnabled = c.Bool(flagIngressEnabled)
	cfg.Ingress.Domain = c.String(flagIngressDomain)
	cfg.Ingress.AnyIPToDomain = c.Bool(flagIngressAnyIPToDomain)
	cfg.Ingress.Namespace = c.String(flagIngressNamespace)
	cfg.Ingress.TLSEnabled = c.Bool(flagIngressTLSEnabled)

	// inference
	cfg.Inference.LogTimeout = c.Duration(flagInferenceLogTimeout)
	cfg.Inference.CacheTTL = c.Duration(flagInferenceCacheTTL)

	// build
	cfg.Build.BuildEnabled = c.Bool(flagBuildEnabled)
	cfg.Build.BuilderImage = c.String(flagBuilderImage)
	cfg.Build.BuildkitdAddress = c.String(flagBuildkitdAddress)
	cfg.Build.BuildCtlBin = c.String(flagBuildCtlBin)
	cfg.Build.BuildRegistry = c.String(flagBuildRegistry)
	cfg.Build.BuildRegistryToken = c.String(flagBuildRegistryToken)
	cfg.Build.BuildImagePullSecret = c.String(flagBuildImagePullSecret)

	// loki
	cfg.Logs.Timeout = c.Duration(flagLogsTimeout)
	cfg.Logs.LokiURL = c.String(flagLogsLokiURL)
	cfg.Logs.LokiUser = c.String(flaglogsLokiUser)
	cfg.Logs.LokiToken = c.String(flagLogsLokiToken)

	// metrics
	cfg.Metrics.PollingInterval = c.Duration(flagMetricsPollingInterval)
	cfg.Metrics.ServerPort = c.Int(flagMetricsPort)
	cfg.Metrics.PrometheusHost = c.String(flagMetricsPrometheusHost)
	cfg.Metrics.PrometheusPort = c.Int(flagMetricsPrometheusPort)

	// postgres database
	cfg.DB.EventEnabled = c.Bool(flagEventEnabled)
	cfg.DB.URL = c.String(flagDBURL)

	// modelz cloud
	cfg.ModelZCloud.Enabled = c.Bool(flagModelZCloudEnabled)
	cfg.ModelZCloud.URL = c.String(flagModelZCloudURL)
	cfg.ModelZCloud.AgentToken = c.String(flagModelZCloudAgentToken)
	cfg.ModelZCloud.HeartbeatInterval = c.Duration(flagModelZCloudAgentHeartbeatInterval)
	cfg.ModelZCloud.Region = c.String(flagModelZCloudRegion)
	cfg.ModelZCloud.UnifiedAPIKey = c.String(flagModelZCloudUnifiedAPIKey)
	cfg.ModelZCloud.UpstreamTimeout = c.Duration(flagModelZCloudUpstreamTimeout)
	cfg.ModelZCloud.MaxIdleConnections = c.Int(flagModelZCloudMaxIdleConnections)
	cfg.ModelZCloud.MaxIdleConnectionsPerHost = c.Int(flagModelZCloudMaxIdleConnectionsPerHost)
	return cfg
}
