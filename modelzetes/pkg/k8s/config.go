// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

// ProbeConfig holds the deployment liveness and readiness options
type ProbeConfig struct {
	InitialDelaySeconds int32
	TimeoutSeconds      int32
	PeriodSeconds       int32
}

// DeploymentConfig holds the global deployment options
type DeploymentConfig struct {
	HTTPProbe                           bool
	ReadinessProbe                      *ProbeConfig
	LivenessProbe                       *ProbeConfig
	StartupProbe                        *ProbeConfig
	HuggingfacePullThroughCache         bool
	HuggingfacePullThroughCacheEndpoint string
	ImagePullPolicy                     string
	// SetNonRootUser will override the function image user to ensure that it is not root. When
	// true, the user will set to 12000 for all functions.
	SetNonRootUser bool
	// ProfilesNamespace defines which namespace is used to look up available Profiles.
	ProfilesNamespace string
}
