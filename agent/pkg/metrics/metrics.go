// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricOptions to be used by web handlers
type MetricOptions struct {
	GatewayInferenceInvocation         *prometheus.CounterVec
	GatewayInferencesHistogram         *prometheus.HistogramVec
	GatewayInferenceInvocationStarted  *prometheus.CounterVec
	GatewayInferenceInvocationInflight *prometheus.GaugeVec

	ServiceReplicasGauge          *prometheus.GaugeVec
	ServiceAvailableReplicasGauge *prometheus.GaugeVec
	ServiceTargetLoad             *prometheus.GaugeVec
}

// ServiceMetricOptions provides RED metrics
type ServiceMetricOptions struct {
	Histogram *prometheus.HistogramVec
	Counter   *prometheus.CounterVec
}

// Synchronize to make sure MustRegister only called once
var once = sync.Once{}

// RegisterExporter registers with Prometheus for tracking
func RegisterExporter(exporter *Exporter) {
	once.Do(func() {
		prometheus.MustRegister(exporter)
	})
}

// PrometheusHandler Bootstraps prometheus for metrics collection
func PrometheusHandler() http.Handler {
	return promhttp.Handler()
}

// BuildMetricsOptions builds metrics for tracking inferences in the API gateway
func BuildMetricsOptions() MetricOptions {
	gatewayInferencesHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "gateway_inferences_seconds",
		Help: "Inference time taken",
	}, []string{"inference_name", "code"})

	gatewayInferenceInvocation := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gateway",
			Subsystem: "inference",
			Name:      "invocation_total",
			Help:      "Inference metrics",
		},
		[]string{"inference_name", "code"},
	)

	serviceReplicas := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gateway",
			Name:      "service_count",
			Help:      "Current count of replicas for inference",
		},
		[]string{"inference_name"},
	)

	serviceAvailableReplicas := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gateway",
			Name:      "service_available_count",
			Help:      "Current count of available replicas for inference",
		},
		[]string{"inference_name"},
	)

	serviceTargetLoad := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gateway",
			Name:      "service_target_load",
			Help:      "Target load for inference",
		},
		[]string{"inference_name", "scaling_type"},
	)

	gatewayInferenceInvocationStarted := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gateway",
			Subsystem: "inference",
			Name:      "invocation_started",
			Help:      "The total number of inference HTTP requests started.",
		},
		[]string{"inference_name"},
	)

	gatewayInferenceInvocationInflight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "gateway",
			Subsystem: "inference",
			Name:      "invocation_inflight",
			Help:      "The number of inference HTTP inflight requests.",
		},
		[]string{"inference_name"},
	)

	metricsOptions := MetricOptions{
		GatewayInferencesHistogram:         gatewayInferencesHistogram,
		GatewayInferenceInvocation:         gatewayInferenceInvocation,
		ServiceReplicasGauge:               serviceReplicas,
		ServiceAvailableReplicasGauge:      serviceAvailableReplicas,
		ServiceTargetLoad:                  serviceTargetLoad,
		GatewayInferenceInvocationStarted:  gatewayInferenceInvocationStarted,
		GatewayInferenceInvocationInflight: gatewayInferenceInvocationInflight,
	}

	return metricsOptions
}
