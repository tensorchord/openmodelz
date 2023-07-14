// Copyright (c) Alex Ellis 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

// ScaleServiceRequest scales the service to the requested replcia count.
type ScaleServiceRequest struct {
	ServiceName  string `json:"serviceName"`
	Replicas     uint64 `json:"replicas"`
	EventMessage string `json:"eventMessage"`
}

// DeleteFunctionRequest delete a deployed function
type DeleteFunctionRequest struct {
	FunctionName string `json:"functionName"`
}
