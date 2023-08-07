package prom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/api/types"
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

func (p PrometheusQuery) AddMetrics(inferences []types.InferenceDeployment) {
	if len(inferences) > 0 {
		ns := inferences[0].Spec.Namespace
		q := fmt.Sprintf(`sum(gateway_inference_invocation_total{inference_name=~".*.%s"}) by (inference_name)`, ns)
		// Restrict query results to only inference names matching namespace suffix.

		results, err := p.Fetch(url.QueryEscape(q))
		if err != nil {
			// log the error but continue, the mixIn will correctly handle the empty results.
			logrus.Debugf("Error querying Prometheus: %s\n", err.Error())
		}
		mixIn(inferences, results)
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

type VectorQueryResponse struct {
	Data struct {
		Result []struct {
			Metric struct {
				Code          string `json:"code"`
				ScalingType   string `json:"scaling_type"`
				InferenceName string `json:"inference_name"`
			}
			Value []interface{} `json:"value"`
		}
	}
}

func mixIn(inferences []types.InferenceDeployment, metrics *VectorQueryResponse) {

	if inferences == nil || metrics == nil {
		return
	}

	for i, inference := range inferences {
		for _, v := range metrics.Data.Result {

			if v.Metric.InferenceName == fmt.Sprintf("%s.%s",
				inference.Spec.Name, inference.Spec.Namespace) {
				metricValue := v.Value[1]
				switch value := metricValue.(type) {
				case string:
					f, err := strconv.ParseFloat(value, 64)
					if err != nil {
						logrus.Debugf("add_metrics: unable to convert value %q for metric: %s", value, err)
						continue
					}
					inferences[i].Status.InvocationCount += int32(f)
				}
			}
		}
	}
}
