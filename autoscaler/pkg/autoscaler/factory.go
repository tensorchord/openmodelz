package autoscaler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/client"

	"github.com/tensorchord/openmodelz/autoscaler/pkg/prom"
)

type Opt struct {
	GatewayHost      string
	PrometheusHost   string
	BasicAuthEnabled bool
	SecretPath       string
	PrometheusPort   int

	Interval time.Duration
}

func New(opt Opt) (*Scaler, error) {
	logrus.Info("Creating autoscaler with options: ", opt)

	gatewayURL, err := url.Parse(opt.GatewayHost)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse gateway host")
	}

	client, err := client.NewClientWithOpts(
		client.WithHost(gatewayURL.String()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	prometheusQuery := prom.NewPrometheusQuery(opt.PrometheusHost, opt.PrometheusPort, &http.Client{})

	as := newScaler(client, &prometheusQuery, newLoadCache(), newInferenceCache())
	return as, nil
}
