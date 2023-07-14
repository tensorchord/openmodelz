package log

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

const (
	// refer to https://grafana.com/docs/loki/latest/api/#query-loki-over-a-range-of-time
	lokiQueryRangePath = "/loki/api/v1/query_range"
)

type RangeQueryResponse struct {
	Data struct {
		Result []struct {
			Stream struct {
				Cluster   string `json:"cluster,omitempty"`
				Container string `json:"container,omitempty"`
				Namespace string `json:"namespace,omitempty"`
				Pod       string `json:"pod,omitempty"`
				Job       string `json:"job,omitempty"`
			}
			Values [][]string `json:"values,omitempty"`
		}
		ResultType string `json:"resultType,omitempty"`
	}
	Status string `json:"status,omitempty"`
}

type LokiAPIRequestor struct {
	client http.Client
	url    string
	user   string
	token  string
}

func NewLokiAPIRequestor(url, user, token string) Requester {
	loki := LokiAPIRequestor{
		url:    url,
		user:   user,
		token:  token,
		client: http.Client{},
	}
	return &loki
}

func (l *LokiAPIRequestor) Query(ctx context.Context, r types.LogRequest) (<-chan types.Message, error) {
	var sinceTime time.Time
	if r.Since != "" {
		var err error
		sinceTime, err = time.Parse(time.RFC3339, r.Since)
		if err != nil {
			return nil, errdefs.InvalidParameter(err)
		}
	}

	logs, err := l.getLogs(ctx, &sinceTime, r.Namespace, r.Name)
	return logs, err
}

func (l *LokiAPIRequestor) getLogs(ctx context.Context, since *time.Time,
	namespace, name string) (<-chan types.Message, error) {
	endpoint, err := url.JoinPath(l.url, lokiQueryRangePath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct the query URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct the Loki request")
	}
	req.SetBasicAuth(l.user, l.token)
	query := url.Values{}
	if since != nil {
		query.Add("start", since.String())
		if time.Since(*since) > time.Hour*24*30 {
			// max query range is 30 days
			query.Add("end", strconv.Itoa(int(since.Add(time.Hour*24*30).UnixNano())))
		}
	}
	query.Add("query", fmt.Sprintf(`{namespace="%s",pod="%s"}`, namespace, name))
	req.URL.RawQuery = query.Encode()
	logrus.Debugf("get log from %s", req.URL.String())

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request the Loki service")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Newf("failed to request the Loki, err[%s]", resp.Status)
	}

	var queryResp RangeQueryResponse
	err = json.NewDecoder(resp.Body).Decode(&queryResp)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}

	if len(queryResp.Data.Result) == 0 {
		return nil, errors.New("result contains ")
	}

	msgStream := make(chan types.Message, LogBufferSize)
	go func() {
		defer close(msgStream)
		for _, value := range queryResp.Data.Result[0].Values {
			timestamp, err := time.Parse(time.RFC3339, value[0])
			if err != nil {
				logrus.Infof("failed to parse timestamp %s during parse log from %s:%s\n",
					value[0], namespace, name)
				continue
			}
			msgStream <- types.Message{
				Timestamp: timestamp,
				Text:      value[1],
				Name:      name,
				Namespace: namespace,
				Instance:  name,
			}
		}
	}()

	return msgStream, nil
}
