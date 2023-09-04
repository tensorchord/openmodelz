package config

import (
	"encoding/json"
	"errors"
	"time"
)

type Config struct {
	Server      ServerConfig      `json:"server,omitempty"`
	KubeConfig  KubeConfig        `json:"kube_config,omitempty"`
	Ingress     IngressConfig     `json:"ingress,omitempty"`
	Inference   InferenceConfig   `json:"inference,omitempty"`
	Build       BuildConfig       `json:"build,omitempty"`
	Metrics     MetricsConfig     `json:"metrics,omitempty"`
	Logs        LogsConfig        `json:"logs,omitempty"`
	DB          PostgresConfig    `json:"db,omitempty"`
	ModelZCloud ModelZCloudConfig `json:"modelz_cloud,omitempty"`
}

type ModelZCloudConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	// URL of apiserver
	URL                       string            `json:"url,omitempty"`
	AgentToken                string            `json:"agent_token,omitempty"`
	HeartbeatInterval         time.Duration     `json:"heartbeat_interval,omitempty"`
	ID                        string            `json:"id,omitempty"`
	TokenID                   string            `json:"token_id,omitempty"`
	Region                    string            `json:"region,omitempty"`
	APIKeys                   map[string]string `json:"api_keys,omitempty"`
	UserNamespaces            []string          `json:"user_namespaces,omitempty"`
	UnifiedAPIKey             string            `json:"unified_api_key,omitempty"`
	UpstreamTimeout           time.Duration     `json:"upstream_timeout,omitempty"`
	MaxIdleConnections        int               `json:"max_idle_connections,omitempty"`
	MaxIdleConnectionsPerHost int               `json:"max_idle_connections_per_host,omitempty"`
}

type LogsConfig struct {
	Timeout   time.Duration `json:"timeout,omitempty"`
	LokiURL   string        `json:"loki_url,omitempty"`
	LokiUser  string        `json:"loki_user,omitempty"`
	LokiToken string        `json:"loki_token,omitempty"`
}

type ServerConfig struct {
	Dev          bool          `json:"dev,omitempty"`
	ServerPort   int           `json:"server_port,omitempty"`
	ReadTimeout  time.Duration `json:"read_timeout,omitempty"`
	WriteTimeout time.Duration `json:"write_timeout,omitempty"`
}

type MetricsConfig struct {
	PollingInterval time.Duration `json:"polling_interval,omitempty"`
	ServerPort      int           `json:"server_port,omitempty"`
	PrometheusPort  int           `json:"prometheus_port,omitempty"`
	PrometheusHost  string        `json:"prometheus_host,omitempty"`
}

type BuildConfig struct {
	BuildEnabled         bool   `json:"build_enabled,omitempty"`
	BuilderImage         string `json:"builder_image,omitempty"`
	BuildkitdAddress     string `json:"buildkitd_address,omitempty"`
	BuildCtlBin          string `json:"build_ctl_bin,omitempty"`
	BuildRegistry        string `json:"build_registry,omitempty"`
	BuildRegistryToken   string `json:"build_registry_token,omitempty"`
	BuildImagePullSecret string `json:"build_image_pull_secret,omitempty"`
}

type InferenceConfig struct {
	LogTimeout time.Duration `json:"log_timeout,omitempty"`
	CacheTTL   time.Duration `json:"cache_ttl,omitempty"`
}

type IngressConfig struct {
	IngressEnabled bool   `json:"ingress_enabled,omitempty"`
	Domain         string `json:"domain,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	AnyIPToDomain  bool   `json:"any_ip_to_domain,omitempty"`
	TLSEnabled     bool   `json:"tls_enabled,omitempty"`
}

type KubeConfig struct {
	Kubeconfig   string        `json:"kubeconfig,omitempty"`
	MasterURL    string        `json:"master_url,omitempty"`
	QPS          int           `json:"qps,omitempty"`
	Burst        int           `json:"burst,omitempty"`
	ResyncPeriod time.Duration `json:"resync_period,omitempty"`
}

type PostgresConfig struct {
	EventEnabled bool   `json:"event_enabled,omitempty"`
	URL          string `json:"url,omitempty"`
}

func New() Config {
	return Config{
		KubeConfig: KubeConfig{},
		Ingress:    IngressConfig{},
		Inference:  InferenceConfig{},
		Build:      BuildConfig{},
		Metrics:    MetricsConfig{},
		Logs:       LogsConfig{},
	}
}

func (c Config) GetString() (string, error) {
	bytes, err := json.Marshal(c)
	return string(bytes), err
}

func (c Config) Validate() error {
	if c.Server.ServerPort == 0 ||
		c.Server.ReadTimeout == 0 ||
		c.Server.WriteTimeout == 0 {
		return errors.New("server config is required")
	}

	if c.Inference.LogTimeout == 0 {
		return errors.New("inference log timeout is required")
	}

	if c.Build.BuildEnabled {
		if c.Build.BuildkitdAddress == "" ||
			c.Build.BuilderImage == "" ||
			c.Build.BuildRegistryToken == "" ||
			c.Build.BuildRegistry == "" ||
			c.Build.BuildCtlBin == "" ||
			c.Build.BuildImagePullSecret == "" {
			return errors.New("build config is required")
		}
	}

	if c.Metrics.ServerPort == 0 ||
		c.Metrics.PollingInterval == 0 ||
		c.Metrics.PrometheusHost == "" ||
		c.Metrics.PrometheusPort == 0 {
		return errors.New("metrics config is required")
	}

	if c.DB.EventEnabled {
		if c.DB.URL == "" {
			return errors.New("db config is required")
		}
	}

	if c.Ingress.IngressEnabled {
		if c.Ingress.Namespace == "" {
			return errors.New("ingress namespace is required")
		}
		if !c.Ingress.AnyIPToDomain && c.Ingress.Domain == "" {
			return errors.New("ingress domain is required")
		}
	}

	if c.ModelZCloud.Enabled {
		if c.ModelZCloud.URL == "" ||
			c.ModelZCloud.AgentToken == "" ||
			c.ModelZCloud.HeartbeatInterval == 0 {
			return errors.New("modelz cloud config is required")
		}
	}
	return nil
}
