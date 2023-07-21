package config

import (
	"encoding/json"
	"errors"
	"time"
)

type Config struct {
	KubeConfig KubeConfig       `json:"kube_config,omitempty"`
	Controller ControllerConfig `json:"controller,omitempty"`
}

type ControllerConfig struct {
	ThreadCount int    `json:"thread_count,omitempty"`
	Namespace   string `json:"namespace,omitempty"`
	Host        string `json:"host,omitempty"`
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

	if c.Controller.ThreadCount == 0 || c.Controller.Namespace == "" ||
		c.Controller.Host == "" {
		return errors.New("invalid controller config")
	}

	return nil
}
