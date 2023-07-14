package server

import (
	"context"

	"github.com/tensorchord/openmodelz/agent/pkg/metrics"
)

func (s *Server) initMetrics() error {
	metricsOptions := metrics.BuildMetricsOptions()
	s.metricsOptions = metricsOptions
	exporter := metrics.NewExporter(metricsOptions, s.runtime)
	metrics.RegisterExporter(exporter)
	exporter.StartServiceWatcher(context.TODO(), s.config.Metrics.PollingInterval)
	return nil
}
