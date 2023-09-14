package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (s *Server) Run() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Server.ServerPort),
		Handler:      s.router,
		WriteTimeout: s.config.Server.WriteTimeout,
		ReadTimeout:  s.config.Server.ReadTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("listen on port %d error: %v", s.config.Server.ServerPort, err)
		}
	}()

	metricsSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Metrics.ServerPort),
		Handler:      s.metricsRouter,
		ReadTimeout:  s.config.Metrics.PollingInterval,
		WriteTimeout: s.config.Metrics.PollingInterval,
	}
	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			logrus.Errorf("listen on port %d error: %v",
				s.config.Metrics.ServerPort, err)
		}
	}()

	logrus.WithField("port", s.config.Server.ServerPort).
		Info("server is running...")
	logrus.WithField("metrics-port", s.config.Metrics.ServerPort).
		Info("metrics server is running...")

	if s.config.ModelZCloud.Enabled {
		// check apiserver is ready
		apiServerReady := make(chan struct{})
		go func() {
			if err := s.modelzCloudClient.WaitForAPIServerReady(); err != nil {
				logrus.Fatalf("failed to wait for apiserver ready: %v", err)
			}
			close(apiServerReady)
		}()
		// websocket
		// build websocket
		go s.connect(apiServerReady)

		// heartbeat with apiserver
		go wait.UntilWithContext(context.Background(), func(ctx context.Context) {
			cluster := types.ManagedCluster{
				ID:        s.config.ModelZCloud.ID,
				Status:    types.ClusterStatusActive,
				UpdatedAt: time.Now().UTC(),
				TokenID:   s.config.ModelZCloud.TokenID,
				Region:    s.config.ModelZCloud.Region,
			}
			err := s.runtime.GetClusterInfo(&cluster)
			if err != nil {
				logrus.Errorf("failed to get managed cluster info: %v", err)
			}

			err = s.modelzCloudClient.UpdateAgentStatus(ctx, apiServerReady, s.config.ModelZCloud.AgentToken, cluster)
			if err != nil {
				logrus.Errorf("failed to update agent status: %v", err)
			}
			logrus.Debugf("update agent status: %v", cluster)
		}, s.config.ModelZCloud.HeartbeatInterval)

		go wait.UntilWithContext(context.Background(), func(ctx context.Context) {
			apikeys, err := s.modelzCloudClient.GetAPIKeys(ctx, apiServerReady, s.config.ModelZCloud.AgentToken, s.config.ModelZCloud.ID)
			if err != nil {
				logrus.Errorf("failed to get apikeys: %v", err)
			}

			s.config.ModelZCloud.APIKeys = apikeys
			logrus.Debugf("update apikeys")
		}, s.config.ModelZCloud.HeartbeatInterval) // default 1min update, TODO(xieydd) make it configurable

		go wait.UntilWithContext(context.Background(), func(ctx context.Context) {
			namespaces, err := s.modelzCloudClient.GetNamespaces(ctx, apiServerReady, s.config.ModelZCloud.AgentToken, s.config.ModelZCloud.ID)
			if err != nil {
				logrus.Errorf("failed to get namespaces: %v", err)
			}

			for _, ns := range namespaces.Items {
				if ContainString(ns, s.config.ModelZCloud.UserNamespaces) {
					continue
				}
				err = s.runtime.NamespaceCreate(ctx, ns)
				if err != nil {
					logrus.Errorf("failed to create namespace %s: %v", ns, err)
					continue
				}
				s.config.ModelZCloud.UserNamespaces = append(s.config.ModelZCloud.UserNamespaces, ns)
				logrus.Debugf("update namespaces")
			}
		}, s.config.ModelZCloud.HeartbeatInterval) // default 1h update, make it configurable
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("shutdown server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func ContainString(target string, strs []string) bool {
	for _, str := range strs {
		if str == target {
			return true
		}
	}
	return false
}
