package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/pkg/server/validator"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/tensorchord/openmodelz/mdz/pkg/agentd/runtime"
)

type Server struct {
	router        *gin.Engine
	metricsRouter *gin.Engine
	logger        *logrus.Entry
	validator     *validator.Validator
	runtime       *runtime.Runtime
}

func New() (*Server, error) {
	router := gin.New()
	router.Use(ginlogrus.Logger(logrus.StandardLogger(), "/healthz"))
	router.Use(gin.Recovery())

	// metrics server
	metricsRouter := gin.New()
	metricsRouter.Use(gin.Recovery())

	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Allow CORS")
		router.Use(cors.New(cors.Config{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
			AllowHeaders: []string{"*"},
		}))
	}

	logger := logrus.WithField("component", "agentd")

	r, err := runtime.New()
	if err != nil {
		return nil, err
	}
	s := &Server{
		router:        router,
		metricsRouter: metricsRouter,
		logger:        logger,
		validator:     validator.New(),
		runtime:       r,
	}

	s.registerRoutes()
	return s, nil
}
