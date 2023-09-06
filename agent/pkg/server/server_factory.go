package server

import (
	"context"
	"net/http"

	"github.com/dgraph-io/ristretto"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/client"
	"github.com/tensorchord/openmodelz/agent/pkg/query"
	ginlogrus "github.com/toorop/gin-logrus"

	"github.com/tensorchord/openmodelz/agent/pkg/config"
	"github.com/tensorchord/openmodelz/agent/pkg/event"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
	"github.com/tensorchord/openmodelz/agent/pkg/log"
	"github.com/tensorchord/openmodelz/agent/pkg/metrics"
	"github.com/tensorchord/openmodelz/agent/pkg/prom"
	"github.com/tensorchord/openmodelz/agent/pkg/runtime"
	"github.com/tensorchord/openmodelz/agent/pkg/scaling"
	"github.com/tensorchord/openmodelz/agent/pkg/server/validator"
)

type Server struct {
	router        *gin.Engine
	metricsRouter *gin.Engine
	logger        *logrus.Entry
	validator     *validator.Validator

	runtime runtime.Runtime

	// endpointResolver resolves the requests from the client to the
	// corresponding inference kubernetes service.
	endpointResolver       k8s.Resolver
	buildLogRequester      log.Requester
	deploymentLogRequester log.Requester

	// prometheusClient is the client to query the prometheus server.
	// It is used in inference list.
	prometheusClient prom.PrometheusQuery
	metricsOptions   metrics.MetricOptions

	// scaler scales the inference from 0 to 1.
	scaler *scaling.InferenceScaler

	config config.Config

	eventRecorder event.Interface

	modelzCloudClient *client.Client

	cache ristretto.Cache
}

func New(c config.Config) (Server, error) {
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

	promCli := prom.NewPrometheusQuery(c.Metrics.PrometheusHost, c.Metrics.PrometheusPort, http.DefaultClient)

	logger := logrus.WithField("component", "server")

	s := Server{
		router:           router,
		metricsRouter:    metricsRouter,
		config:           c,
		logger:           logger,
		validator:        validator.New(),
		prometheusClient: promCli,
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 28,
		BufferItems: 64,
	})
	if err != nil {
		return s, err
	}
	s.cache = *cache

	if s.config.DB.EventEnabled {
		logrus.Info("Event recording is enabled")
		// Connect to database
		conn, err := pgxpool.Connect(context.Background(), c.DB.URL)
		if err != nil {
			return s, errors.Wrap(err, "failed to connect to database")
		}
		queries := query.New(conn)
		s.eventRecorder = event.NewEventRecorder(queries)
	} else {
		s.eventRecorder = event.NewFake()
	}

	s.registerRoutes()
	s.registerMetricsRoutes()
	if err := s.initKubernetesResources(); err != nil {
		return s, err
	}

	if c.ModelZCloud.Enabled {
		err := s.initModelZCloud(c.ModelZCloud.URL, c.ModelZCloud.AgentToken, c.ModelZCloud.Region)
		if err != nil {
			return s, err
		}
	}
	if err := s.initMetrics(); err != nil {
		return s, err
	}
	s.initLogs()
	return s, nil
}
