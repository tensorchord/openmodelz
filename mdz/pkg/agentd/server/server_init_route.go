package server

import (
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	endpointInferencePlural = "/inferences"
	endpointInference       = "/inference"
	endpointScaleInference  = "/scale-inference"
	endpointInfo            = "/info"
	endpointLogPlural       = "/logs"
	endpointNamespacePlural = "/namespaces"
	endpointHealthz         = "/healthz"
	endpointBuild           = "/build"
)

func (s *Server) registerRoutes() {
	root := s.router.Group("/")

	// swagger
	root.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	// dataplane
	root.Any("/inference/:name",
		WrapHandler(s.middlewareCallID),
		WrapHandler(s.handleInferenceProxy))
	root.Any("/inference/:name/*proxyPath",
		WrapHandler(s.middlewareCallID),
		WrapHandler(s.handleInferenceProxy))

	// healthz
	root.GET(endpointHealthz, WrapHandler(s.handleHealthz))

	// control plane
	controlPlane := root.Group("/system")
	// inferences
	controlPlane.GET(endpointInferencePlural,
		WrapHandler(s.handleInferenceList))
	controlPlane.POST(endpointInferencePlural,
		WrapHandler(s.handleInferenceCreate))
	// controlPlane.PUT(endpointInferencePlural,
	// 	WrapHandler(s.handleInferenceUpdate))
	controlPlane.DELETE(endpointInferencePlural,
		WrapHandler(s.handleInferenceDelete))
	// controlPlane.POST(endpointScaleInference,
	// 	WrapHandler(s.handleInferenceScale))
	controlPlane.GET(endpointInference+"/:name",
		WrapHandler(s.handleInferenceGet))

	// instances
	// controlPlane.GET(endpointInference+"/:name/instances",
	// 	WrapHandler(s.handleInferenceInstance))

	// info
	controlPlane.GET(endpointInfo, WrapHandler(s.handleInfo))

	// logs
	controlPlane.GET(endpointLogPlural+endpointInference,
		WrapHandler(s.handleInferenceLogs))
}
