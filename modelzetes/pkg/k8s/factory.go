// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"k8s.io/client-go/kubernetes"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned/typed/modelzetes/v2alpha1"
)

// FunctionFactory is handling Kubernetes operations to materialise functions into deployments and services
type FunctionFactory struct {
	Client kubernetes.Interface
	Config DeploymentConfig
}

func NewFunctionFactory(clientset kubernetes.Interface, config DeploymentConfig, inferenceclientset v2alpha1.InferenceInterface) FunctionFactory {
	return FunctionFactory{
		Client: clientset,
		Config: config,
	}
}
