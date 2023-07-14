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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("shutdown server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}
