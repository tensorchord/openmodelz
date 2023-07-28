package config

import (
	"encoding/json"
	"errors"
	"time"
)

type Config struct {
	Metrics          MetricsConfig          `json:"metrics,omitempty"`
	KubeConfig       KubeConfig             `json:"kube_config,omitempty"`
	Controller       ControllerConfig       `json:"controller,omitempty"`
	HuggingfaceProxy HuggingfaceProxyConfig `json:"huggingface_proxy,omitempty"`
	Probes           ProbesConfig           `json:"probes,omitempty"`
	Inference        InferenceConfig        `json:"inference,omitempty"`
}

type InferenceConfig struct {
	ImagePullPolicy         string `json:"image_pull_policy,omitempty"`
	SetUpRuntimeClassNvidia bool   `json:"set_up_runtime_class_nvidia,omitempty"`
}

type ProbesConfig struct {
	Startup   ProbeConfig `json:"startup,omitempty"`
	Readiness ProbeConfig `json:"readiness,omitempty"`
	Liveness  ProbeConfig `json:"liveness,omitempty"`
}

type ProbeConfig struct {
	InitialDelaySeconds int `json:"initial_delay_seconds,omitempty"`
	PeriodSeconds       int `json:"period_seconds,omitempty"`
	TimeoutSeconds      int `json:"timeout_seconds,omitempty"`
}

type HuggingfaceProxyConfig struct {
	Endpoint string `json:"endpoint,omitempty"`
}

type ControllerConfig struct {
	ThreadCount int `json:"thread_count,omitempty"`
}

type MetricsConfig struct {
	ServerPort int `json:"server_port,omitempty"`
}
type KubeConfig struct {
	Kubeconfig   string        `json:"kubeconfig,omitempty"`
	MasterURL    string        `json:"master_url,omitempty"`
	QPS          int           `json:"qps,omitempty"`
	Burst        int           `json:"burst,omitempty"`
	ResyncPeriod time.Duration `json:"resync_period,omitempty"`
}

func New() Config {
	return Config{}
}

func (c Config) GetString() (string, error) {
	bytes, err := json.Marshal(c)
	return string(bytes), err
}

func (c Config) Validate() error {
	if c.KubeConfig.QPS == 0 ||
		c.KubeConfig.Burst == 0 ||
		c.KubeConfig.ResyncPeriod == 0 {
		return errors.New("invalid kubeconfig")
	}

	// if c.Metrics.ServerPort == 0 {
	// 	return errors.New("invalid metrics config")
	// }

	if c.Controller.ThreadCount == 0 {
		return errors.New("invalid controller config")
	}

	if c.Inference.ImagePullPolicy == "" {
		return errors.New("invalid inference config")
	}
	return nil
}
