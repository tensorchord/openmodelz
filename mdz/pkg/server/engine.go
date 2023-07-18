package server

import (
	"io"
	"time"
)

type Options struct {
	Verbose       bool
	OutputStream  io.Writer
	RetryInternal time.Duration
}

type Engine struct {
	options Options
	Steps   []Step
}

type Result struct {
	AgentURL string
	Command  string
}

func NewStart(o Options) (*Engine, error) {
	return &Engine{
		options: o,
		Steps: []Step{
			// Install k3s and related tools.
			&k3sInstallStep{
				options: o,
			},
			&helmStep{
				options: o,
			},
		},
	}, nil
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

func NewJoin(o Options) (*Engine, error) {
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
	return &Result{
		AgentURL: "http://localhost:31112",
	}, nil
}
