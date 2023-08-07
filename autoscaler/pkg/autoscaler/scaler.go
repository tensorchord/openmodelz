package autoscaler

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/client"
	"github.com/tensorchord/openmodelz/agent/pkg/scaling"

	"github.com/tensorchord/openmodelz/autoscaler/pkg/prom"
)

type Scaler struct {
	PromQuery      *prom.PrometheusQuery
	client         *client.Client
	LoadCache      *LoadCache
	ZeroCache      map[string]time.Time
	InferenceCache *InferenceCache
}

func newScaler(c *client.Client,
	promQuery *prom.PrometheusQuery,
	loadCache *LoadCache,
	inferanceCache *InferenceCache) *Scaler {
	return &Scaler{
		client:         c,
		PromQuery:      promQuery,
		LoadCache:      loadCache,
		ZeroCache:      make(map[string]time.Time),
		InferenceCache: inferanceCache,
	}
}

func (s *Scaler) AutoScale(interval time.Duration) {
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	TTL := 1 * time.Minute

	for {
		select {
		case <-ticker.C:
			// Detect if the instance pod always restart,
			// if pod restart count in 10 minutes before last update time is more than 2, will scale it down.
			results, err := s.GetRestartMetrics()
			if err != nil {
				logrus.Info("Get Restart Metrics of inference Failed")
				continue
			}

			inferenceCount := make(map[string]int)
			for _, ts := range results {
				labels := ts.Labels
				podName, inferenceName, namespace := "", "", ""
				for _, label := range labels {
					switch label.Name {
					case "pod":
						podName = label.Value
					case "inference_name":
						inferenceName = label.Value
					case "namespace":
						namespace = label.Value
					}
				}
				if len(ts.Samples) < 1 {
					logrus.Infof("Sample not found for inference %s.", &inferenceName)
					continue
				}

				strs := strings.Split(inferenceName, ".")
				if len(strs) != 2 {
					logrus.Infof("Invalid inference name: %s", inferenceName)
					continue
				}
				name := strs[0]
				resp, err := s.client.InstanceList(context.TODO(), namespace, name)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"service": inferenceName,
						"error":   err,
					}).Error("failed to get instance list")
					continue
				}

				for _, instance := range resp {
					if instance.Spec.Name == podName {
						if instance.Status.Phase == "CrashLoopBackOff" {
							inferenceCount[inferenceName] += 1
						}
					}
				}
			}

			if len(inferenceCount) != 0 {
				for inferenceName, count := range inferenceCount {
					strs := strings.Split(inferenceName, ".")
					if len(strs) != 2 {
						logrus.Infof("Invalid inference name: %s", inferenceName)
						continue
					}
					name := strs[0]
					namespace := strs[1]

					resp, ok := s.InferenceCache.Get(inferenceName, TTL)
					if !ok {
						resp, err = s.client.InferenceGet(context.TODO(), namespace, name)
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"service": inferenceName,
								"error":   err,
							}).Error("failed to get inference")
							continue
						}

						// update inference cache
						inference := Inference{
							Timestamp:  time.Now(),
							Deployment: resp,
						}
						s.InferenceCache.Set(inferenceName, inference)
					}
					// check if the instance already exists
					var expectedReplicas int
					totalReplicas := resp.Status.Replicas
					if count > int(totalReplicas) {
						expectedReplicas = 0
					} else {
						expectedReplicas = int(totalReplicas) - count
					}

					if expectedReplicas != int(totalReplicas) {
						logrus.Infof("Scaling inference %s to %d replicas", inferenceName, expectedReplicas)
						// Add event to record the scale down operation
						eventMessage := fmt.Sprintf("Deployment %d replicas always CrashLoopBackOff, system scale down the deployment replicas to %d", count, expectedReplicas)
						if err := s.client.InferenceScale(context.TODO(),
							namespace, name, expectedReplicas, eventMessage); err != nil {
							logrus.WithFields(logrus.Fields{
								"service":  inferenceName,
								"expected": expectedReplicas,
								"error":    err,
							}).Error("failed to scale inference")
							continue
						}

						// update the inference, set minReplicas to expectedReplicas
						if resp.Spec.Scaling.MinReplicas != nil &&
							*resp.Spec.Scaling.MinReplicas > int32(expectedReplicas) {
							resp.Status.EventMessage = fmt.Sprintf("Deployment %d replicas always CrashLoopBackOff, system scales down the replicas to %d, original min replicas is %s, reset it to %d",
								count, expectedReplicas, *resp.Spec.Scaling.MinReplicas,
								expectedReplicas)
							*resp.Spec.Scaling.MinReplicas = int32(expectedReplicas)
							if _, err := s.client.DeploymentUpdate(context.TODO(), namespace, resp); err != nil {
								logrus.WithFields(logrus.Fields{
									"service":  inferenceName,
									"expected": expectedReplicas,
									"error":    err,
								}).Error("failed to update inference")
								continue
							}
						}
					}
				}
			}

			s.LoadCache = newLoadCache()
			s.GetLoadMetrics()

			for service, lc := range s.LoadCache.load {
				// if instances of inference are restarting, do not scale it.
				if value, ok := inferenceCount[service]; ok && value > 0 {
					continue
				}
				strs := strings.Split(service, ".")
				if len(strs) != 2 {
					logrus.Infof("Invalid inference name: %s", service)
					continue
				}
				name := strs[0]
				namespace := strs[1]

				resp, ok := s.InferenceCache.Get(service, TTL)
				if !ok {
					resp, err = s.client.InferenceGet(context.TODO(), namespace, name)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"service": service,
							"error":   err,
						}).Error("failed to get inference")
						continue
					}

					// update inference cache
					inference := Inference{
						Timestamp:  time.Now(),
						Deployment: resp,
					}
					s.InferenceCache.Set(service, inference)
				}

				if resp.Spec.Labels == nil {
					logrus.WithFields(logrus.Fields{
						"service": service,
						"error":   err,
					}).Error("failed to get inference labels")
					continue
				}

				var expectedReplicas int
				var targetLoad int
				// If the inference has a target load label, use that instead.
				if resp.Spec.Scaling != nil && resp.Spec.Scaling.TargetLoad != nil {
					targetLoad = int(*resp.Spec.Scaling.TargetLoad)
					expectedReplicas = int(math.Ceil(
						lc.CurrentLoad / float64(*resp.Spec.Scaling.TargetLoad)))
				}

				if expectedReplicas == 0 {
					// Check the current start requests to see if the inference is being used.
					if lc.CurrentStartedRequests > 0 {
						logrus.WithFields(logrus.Fields{
							"service":                  service,
							"current_started_requests": lc.CurrentStartedRequests,
							"target_load":              lc.CurrentLoad,
						}).Debug("inference is being used")
						expectedReplicas = 1
					}
				}

				var maxReplicas, minReplicas int
				var zeroDuration time.Duration
				if resp.Spec.Scaling != nil {
					if resp.Spec.Scaling.MinReplicas != nil {
						minReplicas = int(*resp.Spec.Scaling.MinReplicas)
					} else {
						minReplicas = scaling.DefaultMinReplicas
					}

					if resp.Spec.Scaling.MaxReplicas != nil {
						maxReplicas = int(*resp.Spec.Scaling.MaxReplicas)
					} else {
						maxReplicas = scaling.DefaultMaxReplicas
					}

					if resp.Spec.Scaling.ZeroDuration != nil {
						zeroDuration = time.Duration(*resp.Spec.Scaling.ZeroDuration) * time.Second
					} else {
						zeroDuration = scaling.DefaultZeroDuration
					}
				}

				if expectedReplicas > maxReplicas {
					logrus.Infof("Expected replicas (%d) exceeds max replicas (%d) for inference %s", expectedReplicas, maxReplicas, service)
					expectedReplicas = maxReplicas
				}
				if expectedReplicas < minReplicas {
					logrus.Infof("Expected replicas (%d) is less than min replicas (%d) for inference %s", expectedReplicas, minReplicas, service)
					expectedReplicas = minReplicas
				}

				availableReplicas := resp.Status.AvailableReplicas
				totalReplicas := resp.Status.Replicas

				if expectedReplicas == int(totalReplicas) {
					// If the expected replicas is the same as the current replicas, remove the entry from the zero cache.
					delete(s.ZeroCache, service)
					logrus.WithFields(logrus.Fields{
						"service":          service,
						"replicas":         totalReplicas,
						"expectedReplicas": expectedReplicas,
					}).Debug("delete zero cache")
				}

				if expectedReplicas == 0 && totalReplicas != 0 {
					if availableReplicas == 0 {
						// If the expected replicas is 0 and there are no available replicas,
						// set the expected replicas to 1 to prevent the inference from being scaled to zero.
						expectedReplicas = 1
					} else {
						// If the expected replicas is 0 and there is no entry in the zero cache, add one.
						if _, ok := s.ZeroCache[service]; !ok {
							s.ZeroCache[service] = time.Now()
						}

						// If the inference has been idle for longer than the zero duration, scale to zero.
						if time.Since(s.ZeroCache[service]) > zeroDuration {
							logrus.Infof("Inference %s has been idle for %s, scaling to zero", service, zeroDuration)
						} else {
							// If the inference has not been idle for longer than the zero duration, scale to 1.
							expectedReplicas = 1
						}
					}
				}

				if expectedReplicas == 1 && totalReplicas == 0 {
					// If the expected replicas is 1 and the current replicas is 0, do nothing since the scaling handler in gateway will take care of this situation.
					expectedReplicas = 0
				}

				logrus.WithFields(logrus.Fields{
					"service":           service,
					"replicas":          totalReplicas,
					"expectedReplicas":  expectedReplicas,
					"availableReplicas": availableReplicas,
					"currentLoad":       lc.CurrentLoad,
					"targetLoad":        targetLoad,
					"zeroDuration":      zeroDuration,
					"zeroCache":         s.ZeroCache[service],
				}).Debug("start scaling (replicas)")

				if expectedReplicas != int(totalReplicas) {
					delete(s.ZeroCache, service)
					logrus.Infof("Scaling inference %s to %d replicas", service, expectedReplicas)
					eventMessage := fmt.Sprintf("Scaling inference based load, current %f, target %d",
						lc.CurrentLoad, targetLoad)
					if err := s.client.InferenceScale(context.TODO(),
						namespace, name, expectedReplicas, eventMessage); err != nil {
						logrus.WithFields(logrus.Fields{
							"service":  service,
							"expected": expectedReplicas,
							"error":    err,
						}).Error("failed to scale inference")
						continue
					}
				}
			}
		case <-quit:
			return
		}
	}
}

func (s *Scaler) GetLoadMetrics() {
	results, err := s.PromQuery.Fetch(url.QueryEscape("job:inference_current_load:sum"))
	if err != nil {
		// log the error but continue, the mixIn will correctly handle the empty results.
		logrus.Infof("Error querying Prometheus: %s\n", err.Error())
	}

	currentSumResults, err := s.PromQuery.Fetch(
		url.QueryEscape("job:inference_current_started:max_sum"))
	if err != nil {
		// log the error but continue, the mixIn will correctly handle the empty results.
		logrus.Infof("Error querying Prometheus: %s\n", err.Error())
	}

	for _, result := range results.Data.Result {
		currentLoad := 0.0

		switch val := result.Value[1].(type) {
		case string:
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				logrus.Infof("add_metrics: unable to convert value %q for metric: %s", val, err)
				continue
			}
			currentLoad = f
		}

		timestamp := time.Now()
		switch val := result.Value[0].(type) {
		case float64:
			timestamp = time.Unix(int64(val), 0)
		}

		if l, ok := s.LoadCache.Get(result.Metric.InferenceName); ok {
			l.CurrentLoad = currentLoad
			l.Timestamp = timestamp
			s.LoadCache.Set(result.Metric.InferenceName, l)
		} else {
			s.LoadCache.Set(result.Metric.InferenceName, Load{
				CurrentLoad: currentLoad,
				Timestamp:   timestamp,
			})
		}
	}

	for _, result := range currentSumResults.Data.Result {
		currentSum := 0.0

		switch val := result.Value[1].(type) {
		case string:
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				logrus.Infof("add_metrics: unable to convert value %q for metric: %s", val, err)
				continue
			}
			currentSum = f
		}

		timestamp := time.Now()
		switch val := result.Value[0].(type) {
		case float64:
			timestamp = time.Unix(int64(val), 0)
		}

		if l, ok := s.LoadCache.Get(result.Metric.InferenceName); ok {
			l.CurrentStartedRequests = currentSum
			l.Timestamp = timestamp
			s.LoadCache.Set(result.Metric.InferenceName, l)
		} else {
			s.LoadCache.Set(result.Metric.InferenceName, Load{
				CurrentStartedRequests: currentSum,
				Timestamp:              timestamp,
			})
		}
	}
}

func (s *Scaler) GetRestartMetrics() ([]*prom.TimeSeries, error) {
	// record this rule in prometheus
	// (sum by (pod,namespace) (increase(kube_pod_container_status_restarts_total{namespace=~"modelz-(.*)"}[10m])) > 2) * on (pod) group_left(inference_name) (label_join(label_replace(kube_pod_info{created_by_kind="ReplicaSet",namespace=~"modelz-(.*)"}, "inference", "$1", "created_by_name", "(.+)-.+"), "inference_name",".","inference","namespace"))
	query := "pod_restart_count_over_2_10m"
	tsList, err := s.PromQuery.Query(query, time.Now())
	if err != nil {
		logrus.Infof("Error querying Prometheus: %s\n", err.Error())
		return nil, err
	}

	return tsList, nil
}
