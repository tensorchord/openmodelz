package prom

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	promapiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	prommodel "github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
)

// PrometheusQuery represents parameters for querying Prometheus
type PrometheusQuery struct {
	Port   int
	Host   string
	Client *http.Client
}

type PrometheusQueryFetcher interface {
	Fetch(query string) (*VectorQueryResponse, error)
}

// NewPrometheusQuery create a NewPrometheusQuery
func NewPrometheusQuery(host string, port int, client *http.Client) PrometheusQuery {
	return PrometheusQuery{
		Client: client,
		Host:   host,
		Port:   port,
	}
}

// Fetch queries aggregated stats
func (q PrometheusQuery) Fetch(query string) (*VectorQueryResponse, error) {

	req, reqErr := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%d/api/v1/query?query=%s", q.Host, q.Port, query), nil)
	if reqErr != nil {
		return nil, reqErr
	}

	res, getErr := q.Client.Do(req)
	if getErr != nil {
		return nil, getErr
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code from Prometheus want: %d, got: %d, body: %s", http.StatusOK, res.StatusCode, string(bytesOut))
	}

	var values VectorQueryResponse

	unmarshalErr := json.Unmarshal(bytesOut, &values)
	if unmarshalErr != nil {
		return nil, fmt.Errorf("error unmarshalling result: %s, '%s'", unmarshalErr, string(bytesOut))
	}

	return &values, nil
}

// TODO(xieydd) Refactor PrometheusQuery
// Query queries Prometheus with given query string and time
func (q PrometheusQuery) Query(query string, time time.Time) ([]*TimeSeries, error) {
	var ts []*TimeSeries
	client, err := api.NewClient(api.Config{
		Address: fmt.Sprintf("http://%s:%d", q.Host, q.Port),
	})
	if err != nil {
		return ts, err
	}

	api := promapiv1.NewAPI(client)
	results, warnings, err := api.Query(context.TODO(), query, time)
	if len(warnings) != 0 {
		logrus.Info("Prom query warnings", "warnings", warnings)
	}
	if err != nil {
		return ts, err
	}
	logrus.Info("Prom query result", "result", results.String(), "resultsType", results.Type())
	return convertPromResultsToTimeSeries(results)
}

func convertPromResultsToTimeSeries(value prommodel.Value) ([]*TimeSeries, error) {
	var results []*TimeSeries
	typeValue := value.Type()
	switch typeValue {
	case prommodel.ValMatrix:
		if matrix, ok := value.(prommodel.Matrix); ok {
			for _, sampleStream := range matrix {
				if sampleStream == nil {
					continue
				}
				ts := NewTimeSeries()
				for key, val := range sampleStream.Metric {
					ts.AppendLabel(string(key), string(val))
				}
				for _, pair := range sampleStream.Values {
					ts.AppendSample(int64(pair.Timestamp/1000), float64(pair.Value))
				}
				results = append(results, ts)
			}
			return results, nil
		} else {
			return results, fmt.Errorf("prometheus value type is %v, but assert failed", typeValue)
		}

	case prommodel.ValVector:
		if vector, ok := value.(prommodel.Vector); ok {
			for _, sample := range vector {
				if sample == nil {
					continue
				}
				ts := NewTimeSeries()
				for key, val := range sample.Metric {
					ts.AppendLabel(string(key), string(val))
				}
				// for vector, all the sample has the same timestamp. just one point for each metric
				ts.AppendSample(int64(sample.Timestamp/1000), float64(sample.Value))
				results = append(results, ts)
			}
			return results, nil
		} else {
			return results, fmt.Errorf("prometheus value type is %v, but assert failed", typeValue)
		}
	case prommodel.ValScalar:
		return results, fmt.Errorf("not support for scalar when use timeseries")
	case prommodel.ValString:
		return results, fmt.Errorf("not support for string when use timeseries")
	case prommodel.ValNone:
		return results, fmt.Errorf("prometheus return value type is none")
	}
	return results, fmt.Errorf("prometheus return unknown model value type %v", typeValue)
}
