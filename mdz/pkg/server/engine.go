package server

import (
	"fmt"
	"io"
	"time"
)

const (
	AgentPort = 31112
)

type Options struct {
	Verbose       bool
	OutputStream  io.Writer
	Runtime       Runtime
	Mirror        Mirror
	RetryInternal time.Duration
	ServerIP      string
	Domain        *string
	Version       string
	ForceGPU      bool
}

type Mirror struct {
	Name      string
	Endpoints []string
}

func (m *Mirror) Configured() bool {
	return m.Name != "" && len(m.Endpoints) > 0
}

type Runtime string

var (
	RuntimeK3s    Runtime = "k3s"
	RuntimeDocker Runtime = "docker"
)

type Engine struct {
	options Options
	Steps   []Step
}

type Result struct {
	MDZURL string
}

func NewStart(o Options) (*Engine, error) {
	if o.Verbose {
		fmt.Fprintf(o.OutputStream, "Starting the server with config: %+v\n", o)
	}
	var engine *Engine
	switch o.Runtime {
	case RuntimeDocker:
		engine = &Engine{
			options: o,
			Steps: []Step{
				&agentDRunStep{
					options: o,
				},
			},
		}
	default:
		engine = &Engine{
			options: o,
			Steps: []Step{
				// Install k3s and related tools.
				&k3sPrepare{
					options: o,
				},
				&k3sInstallStep{
					options: o,
				},
				&nginxInstallStep{
					options: o,
				},
				&gpuInstallStep{
					options: o,
				},
				&openModelZInstallStep{
					options: o,
				},
			},
		}
	}
	return engine, nil
}

func NewStop(o Options) (*Engine, error) {
	return &Engine{
		options: o,
		Steps: []Step{
			// Kill all k3s and related tools.
			&k3sKillAllStep{
				options: o,
			},
		},
	}, nil
}

func NewDestroy(o Options) (*Engine, error) {
	return &Engine{
		options: o,
		Steps: []Step{
			// Destroy all k3s and related tools.
			&k3sDestroyAllStep{
				options: o,
			},
		},
	}, nil
}

func NewJoin(o Options) (*Engine, error) {
	return &Engine{
		options: o,
		Steps: []Step{
			&k3sJoinStep{
				options: o,
			},
		},
	}, nil
}

type Step interface {
	Run() error
	Verify() error
}

func (e *Engine) Run() (*Result, error) {
	for _, step := range e.Steps {
		if err := step.Run(); err != nil {
			return nil, err
		}
		// Retry until verify success.
		ticker := time.NewTicker(e.options.RetryInternal)
		for range ticker.C {
			if err := step.Verify(); err == nil {
				ticker.Stop()
				break
			}
		}
	}
	if e.options.Domain != nil {
		return &Result{
			MDZURL: fmt.Sprintf("http://%s", *e.options.Domain),
		}, nil
	}
	// Get the server IP.
	if resultDomain != "" {
		return &Result{
			MDZURL: fmt.Sprintf("http://%s", resultDomain),
		}, nil
	}
	return &Result{
		MDZURL: fmt.Sprintf("http://0.0.0.0:%d", AgentPort),
	}, nil
}
