package autoscaler

import "time"

// LoadCache is a cache for load metrics.
type LoadCache struct {
	load map[string]Load
}

type Load struct {
	ScalingType            string
	CurrentStartedRequests float64
	CurrentLoad            float64
	Timestamp              time.Time
}

func newLoadCache() *LoadCache {
	return &LoadCache{
		load: make(map[string]Load),
	}
}

func (l *LoadCache) Get(key string) (Load, bool) {
	load, ok := l.load[key]
	return load, ok
}

func (l *LoadCache) Set(key string, load Load) {
	l.load[key] = load
}
