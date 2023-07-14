// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type FunctionProbes struct {
	Liveness  *corev1.Probe
	Readiness *corev1.Probe
	Startup   *corev1.Probe
}

// MakeProbes returns the liveness and readiness probes
// by default the health check runs `cat /tmp/.lock` every ten seconds
func (f *FunctionFactory) MakeProbes(port int, httpProbePath string) (*FunctionProbes, error) {
	var handler corev1.ProbeHandler

	if f.Config.HTTPProbe {
		handler = corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: httpProbePath,
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: int32(port),
				},
			},
		}
	} else {
		return nil, nil
	}

	probes := FunctionProbes{}
	probes.Readiness = &corev1.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: f.Config.ReadinessProbe.InitialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.ReadinessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.ReadinessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	probes.Liveness = &corev1.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: f.Config.LivenessProbe.InitialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.LivenessProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.LivenessProbe.PeriodSeconds),
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}

	probes.Startup = &corev1.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: f.Config.StartupProbe.InitialDelaySeconds,
		TimeoutSeconds:      int32(f.Config.StartupProbe.TimeoutSeconds),
		PeriodSeconds:       int32(f.Config.StartupProbe.PeriodSeconds),
		SuccessThreshold:    1,
		// Set failure threshold to 30 to allow for slow-starting inferences.
		FailureThreshold: 30,
	}

	return &probes, nil
}
