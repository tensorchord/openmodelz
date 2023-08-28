package scaling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/sirupsen/logrus"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/runtime"
)

const (
	maxPollCount = 1000
	retries      = 20
	pollInterval = time.Millisecond * 100
)

// InferenceScaler create a new scaler with the specified
// ScalingConfig
func NewInferenceScaler(r runtime.Runtime,
	defaultTTL time.Duration) (*InferenceScaler, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 28,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}
	return &InferenceScaler{
		cache:      *cache,
		runtime:    r,
		defaultTTL: defaultTTL,
	}, nil
}

// InferenceScaler scales from zero
type InferenceScaler struct {
	cache   ristretto.Cache
	mu      sync.RWMutex
	runtime runtime.Runtime

	defaultTTL time.Duration
}

// FunctionScaleResult holds the result of scaling from zero
type FunctionScaleResult struct {
	Available bool
	Error     error
	Found     bool
	Duration  time.Duration
}

func (s *InferenceScaler) get(
	namespace, inferenceName string) (ServiceQueryResponse, error) {
	key := inferenceName + "." + namespace

	s.mu.RLock()
	raw, exit := s.cache.Get(key)
	s.mu.RUnlock()
	if exit {
		return raw.(ServiceQueryResponse), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	raw, exit = s.cache.Get(key)
	if exit {
		return raw.(ServiceQueryResponse), nil
	}

	// The wasn't a hit, or there were no available replicas found
	// so query the live endpoint
	inf, err := s.runtime.InferenceGet(namespace, inferenceName)
	if err != nil {
		return ServiceQueryResponse{}, err
	}
	sqr, err := AsServerQueryResponse(inf)
	if err != nil {
		return ServiceQueryResponse{}, err
	}
	if sqr == nil {
		return ServiceQueryResponse{},
			fmt.Errorf("unable to get service query response")
	}
	s.cache.SetWithTTL(key, *sqr, 1, s.defaultTTL)
	return *sqr, nil
}

// Scale scales a function from zero replicas to 1 or the value set in
// the minimum replicas metadata
func (s *InferenceScaler) Scale(ctx context.Context,
	namespace, inferenceName string) FunctionScaleResult {
	start := time.Now()

	resp, err := s.get(namespace, inferenceName)
	if err != nil {
		return FunctionScaleResult{
			Error:     err,
			Available: false,
			Found:     false,
			Duration:  time.Since(start),
		}
	}

	// Check if there are available replicas in the live data
	if resp.AvailableReplicas > 0 {
		return FunctionScaleResult{
			Error:     nil,
			Available: true,
			Found:     true,
			Duration:  time.Since(start),
		}
	}

	// If the desired replica count is 0, then a scale up event
	// is required.
	if resp.Replicas == 0 {
		// If the max replicas is 0, then the function is not
		// scalable
		if resp.MaxReplicas == 0 {
			return FunctionScaleResult{
				Error:     fmt.Errorf("unable to scale up %s, max replicas is 0", inferenceName),
				Available: false,
				Found:     true,
				Duration:  time.Since(start),
			}
		}

		minReplicas := uint64(1)
		if resp.MinReplicas > 0 {
			minReplicas = resp.MinReplicas
		}

		// In a retry-loop, first query desired replicas, then
		// set them if the value is still at 0.
		scaleResult := Retry(func(attempt int) error {
			inf, err := s.runtime.InferenceGet(namespace, inferenceName)
			if err != nil {
				return err
			}

			// The scale up is complete because the desired replica count
			// has been set to 1 or more.
			if inf.Status.Replicas > 0 {
				return nil
			}

			// Request a scale up to the minimum amount of replicas
			if err := s.runtime.InferenceScale(ctx, namespace, types.ScaleServiceRequest{
				ServiceName:  inferenceName,
				Replicas:     minReplicas,
				EventMessage: fmt.Sprintf("scale up to minimum replicas %d", minReplicas),
			}, inf); err != nil {
				return err
			}
			logrus.WithField("inference", inferenceName).
				WithField("replicas", minReplicas).
				Debug("scaling up inference")
			return nil
		}, "Scale", retries, pollInterval)

		if scaleResult != nil {
			return FunctionScaleResult{
				Error:     scaleResult,
				Available: false,
				Found:     true,
				Duration:  time.Since(start),
			}
		}
	}

	switch resp.Framework {
	// Return early for prototype frameworks.
	case "gradio", "streamlit":
		return FunctionScaleResult{
			Error:     nil,
			Available: false,
			Found:     true,
			Duration:  time.Since(start),
		}
	}

	// Holding pattern for at least one function replica to be available
	for i := 0; i < maxPollCount; i++ {
		inf, err := s.runtime.InferenceGet(namespace, inferenceName)
		if err != nil {
			return FunctionScaleResult{
				Error:     err,
				Available: false,
				Found:     true,
				Duration:  time.Since(start),
			}
		}

		totalTime := time.Since(start)
		if inf.Status.AvailableReplicas > 0 {
			logrus.Debugf("[Ready] function=%s waited for - %.4fs",
				inferenceName, totalTime.Seconds())

			return FunctionScaleResult{
				Error:     nil,
				Available: true,
				Found:     true,
				Duration:  totalTime,
			}
		}

		time.Sleep(pollInterval)
	}

	return FunctionScaleResult{
		Error:     nil,
		Available: true,
		Found:     true,
		Duration:  time.Since(start),
	}
}
