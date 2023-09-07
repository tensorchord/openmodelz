package server

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (s *Server) connect(apiServerReady <-chan struct{}) {
	<-apiServerReady
	var clusterDialEndpoint string
	headers := http.Header{
		"X-Cluster-ID": {s.config.ModelZCloud.ID},
		"Agent-Token":  {s.config.ModelZCloud.AgentToken},
	}
	u, err := url.Parse(s.config.ModelZCloud.URL)
	if err != nil {
		logrus.Errorf("failed to parse url: %v", err)
	}

	switch u.Scheme {
	case "http":
		clusterDialEndpoint = "ws://" + u.Host + types.DailEndPointSuffix
	case "https":
		clusterDialEndpoint = "wss://" + u.Host + types.DailEndPointSuffix
	}

	ctx := context.Background()
	go func() {
		for {
			remotedialer.ClientConnect(ctx, clusterDialEndpoint, headers, nil, func(proto, address string) bool { return true }, nil)
			select {
			case <-ctx.Done():
				return
			case <-time.After(s.config.ModelZCloud.HeartbeatInterval):
				// retry connect after interval
			}
		}
	}()

	// retry(
	// 	s.config.ModelZCloud.HeartbeatInterval,
	// 	func() error {
	// 		logrus.Debugf("run websocket server")
	// 		ctx := context.Background()
	// 		err := remotedialer.ClientConnect(ctx, clusterDialEndpoint, headers, nil,
	// 			func(proto, address string) bool { return true }, nil)
	// 		if err != nil {
	// 			logrus.Errorf("failed to connect to apiserver: %v", err)
	// 			return err
	// 		}
	// 		return nil
	// 	},
	// )

}

func retry(sleep time.Duration, f func() error) {
	i := 1
	for {
		err := f()
		if err == nil {
			return
		} else {
			logrus.Errorf("retry %d times, still failed", i)
			time.Sleep(sleep)
			i++
		}
	}
}
