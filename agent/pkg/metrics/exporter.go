// Copyright (c) Alex Ellis 2017
// Copyright (c) 2018 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/runtime"
)

// Exporter is a prometheus metrics collector.
// It is an implementation of https://pkg.go.dev/github.com/prometheus/client_golang/prometheus#Collector.
type Exporter struct {
	metricOptions MetricOptions
	runtime       runtime.Runtime
	services      []types.InferenceDeployment
	logger        *logrus.Entry
}

// NewExporter creates a new exporter for the OpenFaaS gateway metrics
func NewExporter(options MetricOptions, r runtime.Runtime) *Exporter {
	return &Exporter{
		metricOptions: options,
		runtime:       r,
		services:      []types.InferenceDeployment{},
		logger:        logrus.WithField("component", "exporter"),
	}
}

// Describe is to describe the metrics for Prometheus
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	e.metricOptions.GatewayInferenceInvocation.Describe(ch)
	e.metricOptions.GatewayInferencesHistogram.Describe(ch)
	e.metricOptions.ServiceReplicasGauge.Describe(ch)
	e.metricOptions.ServiceAvailableReplicasGauge.Describe(ch)
	e.metricOptions.ServiceTargetLoad.Describe(ch)
	e.metricOptions.GatewayInferenceInvocationStarted.Describe(ch)
	e.metricOptions.GatewayInferenceInvocationInflight.Describe(ch)
}

// Collect collects data to be consumed by prometheus
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.metricOptions.GatewayInferenceInvocation.Collect(ch)
	e.metricOptions.GatewayInferencesHistogram.Collect(ch)

	e.metricOptions.ServiceReplicasGauge.Reset()
	e.metricOptions.ServiceAvailableReplicasGauge.Reset()
	e.metricOptions.ServiceTargetLoad.Reset()

	for _, service := range e.services {
		var serviceName string
		if len(service.Spec.Namespace) > 0 {
			serviceName = fmt.Sprintf("%s.%s", service.Spec.Name,
				service.Spec.Namespace)
		} else {
			serviceName = service.Spec.Name
		}
		// Initial services information if nil after recent deployment
		e.metricOptions.GatewayInferenceInvocationStarted.WithLabelValues(serviceName)
		e.metricOptions.GatewayInferenceInvocationInflight.WithLabelValues(serviceName)
		// Set current replica count
		e.metricOptions.ServiceReplicasGauge.WithLabelValues(serviceName).
			Set(float64(service.Status.Replicas))
		// Set available replica count
		e.metricOptions.ServiceAvailableReplicasGauge.WithLabelValues(serviceName).
			Set(float64(service.Status.AvailableReplicas))

		// Set target load
		if service.Spec.Scaling != nil {
			e.metricOptions.ServiceTargetLoad.WithLabelValues(
				serviceName, string(*service.Spec.Scaling.Type)).
				Set(float64(*service.Spec.Scaling.TargetLoad))
		}
	}

	e.metricOptions.GatewayInferenceInvocationStarted.Collect(ch)
	e.metricOptions.GatewayInferenceInvocationInflight.Collect(ch)
	e.metricOptions.ServiceReplicasGauge.Collect(ch)
	e.metricOptions.ServiceAvailableReplicasGauge.Collect(ch)
	e.metricOptions.ServiceTargetLoad.Collect(ch)
}

// StartServiceWatcher starts a ticker and collects service replica counts to expose to prometheus
func (e *Exporter) StartServiceWatcher(
	ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:

				namespaces, err := e.runtime.NamespaceList(ctx)
				if err != nil {
					e.logger.Debug("unable to list namespaces: ", err)
				}

				services := []types.InferenceDeployment{}

				for _, namespace := range namespaces {
					nsServices, err := e.runtime.InferenceList(namespace)
					if err != nil {
						e.logger.Debug("unable to list services: ", err)
						continue
					}
					services = append(services, nsServices...)
				}

				e.services = services
				break
			case <-quit:
				return
			}
		}
	}()
}
