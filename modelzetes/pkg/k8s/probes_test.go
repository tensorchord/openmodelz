// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"testing"
)

func Test_makeProbes_useHTTPProbe(t *testing.T) {
	f := mockFactory()
	f.Config.HTTPProbe = true

	probes, err := f.MakeProbes(8080, "/")
	if err != nil {
		t.Fatal(err)
	}

	if probes.Readiness.HTTPGet == nil {
		t.Errorf("Readiness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Liveness.HTTPGet == nil {
		t.Errorf("Liveness probe should have had HTTPGet handler")
		t.Fail()
	}
}

func Test_makeProbes_useCustomDurationHTTPProbe(t *testing.T) {
	f := mockFactory()
	f.Config.HTTPProbe = true
	f.Config.LivenessProbe = &ProbeConfig{
		PeriodSeconds:       1,
		TimeoutSeconds:      3,
		InitialDelaySeconds: 0,
	}
	f.Config.ReadinessProbe = &ProbeConfig{
		PeriodSeconds:       1,
		TimeoutSeconds:      3,
		InitialDelaySeconds: 0,
	}
	f.Config.StartupProbe = &ProbeConfig{
		PeriodSeconds:       1,
		TimeoutSeconds:      3,
		InitialDelaySeconds: 0,
	}

	customDelay := "0"

	probes, err := f.MakeProbes(8080, "/")
	if err != nil {
		t.Fatal(err)
	}

	if probes.Readiness.HTTPGet == nil {
		t.Errorf("Readiness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Readiness.InitialDelaySeconds != 0 {
		t.Errorf("Readiness probe should have initial delay seconds set to %s", customDelay)
		t.Fail()
	}

	if probes.Liveness.HTTPGet == nil {
		t.Errorf("Liveness probe should have had HTTPGet handler")
		t.Fail()
	}
	if probes.Liveness.InitialDelaySeconds != 0 {
		t.Errorf("Readiness probe should have had HTTPGet handler set to %s", customDelay)
		t.Fail()
	}
}
