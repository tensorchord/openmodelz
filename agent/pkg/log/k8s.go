package log

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/internalinterfaces"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

const (
	// podInformerResync is the period between cache syncs in the pod informer
	podInformerResync = 5 * time.Second

	// defaultLogSince is the fallback log stream history
	defaultLogSince = 5 * time.Minute

	// LogBufferSize number of log messages that may be buffered
	LogBufferSize = 500 * 2
)

// K8sAPIRequestor implements the Requestor interface for k8s
type K8sAPIRequestor struct {
	client kubernetes.Interface
}

func NewK8sAPIRequestor(client kubernetes.Interface) Requester {
	return &K8sAPIRequestor{
		client: client,
	}
}

func (k *K8sAPIRequestor) Query(ctx context.Context,
	r types.LogRequest) (<-chan types.Message, error) {
	var sinceTime, endTime time.Time
	if r.Since != "" {
		var err error
		sinceTime, err = time.Parse(time.RFC3339, r.Since)
		if err != nil {
			return nil, errdefs.InvalidParameter(err)
		}
	}

	if r.End != "" {
		var err error
		endTime, err = time.Parse(time.RFC3339, r.End)
		if err != nil {
			return nil, errdefs.InvalidParameter(err)
		}
	} else if r.Follow {
		// avoid truncate
		endTime = time.Now().Add(time.Hour)
	} else {
		endTime = time.Now()
	}

	logStream, err := getLogs(ctx,
		k.client, r.Name, r.Namespace, int64(r.Tail), &sinceTime, r.Follow)
	if err != nil {
		return nil, err
	}

	msgStream := make(chan types.Message, LogBufferSize)
	go func() {
		defer close(msgStream)
		// here we depend on the fact that logStream will close when the context is cancelled,
		// this ensures that the go routine will resolve
		for msg := range logStream {
			// if we have an end time, we should stop streaming logs after that time
			if endTime.After(msg.Timestamp) {
				msgStream <- types.Message{
					Timestamp: msg.Timestamp,
					Text:      msg.Text,
					Name:      msg.Name,
					Instance:  msg.Instance,
					Namespace: msg.Namespace,
				}
			}
		}
	}()

	return msgStream, nil
}

// getLogs returns a channel of logs for the given function
func getLogs(ctx context.Context, client kubernetes.Interface, functionName,
	namespace string, tail int64, since *time.Time, follow bool) (
	<-chan types.Message, error) {
	added, err := startFunctionPodInformer(ctx, client, functionName, namespace)
	if err != nil {
		return nil, err
	}

	logs := make(chan types.Message, LogBufferSize)

	go func() {
		var watching uint
		defer close(logs)

		finished := make(chan error)

		for {
			select {
			case <-ctx.Done():
				return
			case <-finished:
				watching--
				if watching == 0 && !follow {
					return
				}
			case p := <-added:
				watching++
				go func() {
					finished <- podLogs(ctx, client.CoreV1().Pods(namespace),
						p, functionName, namespace, tail, since, follow, logs)
				}()
			}
		}
	}()

	return logs, nil
}

// podLogs returns a stream of logs lines from the specified pod
func podLogs(ctx context.Context, i v1.PodInterface, pod, container,
	namespace string, tail int64, since *time.Time, follow bool,
	dst chan<- types.Message) error {
	opts := &corev1.PodLogOptions{
		Follow:     follow,
		Timestamps: true,
		Container:  container,
	}

	if tail > 0 {
		opts.TailLines = &tail
	}

	if opts.TailLines == nil || since != nil {
		opts.SinceSeconds = parseSince(since)
	}

	stream, err := i.GetLogs(pod, opts).Stream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	done := make(chan error)
	go func() {
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			msg, ts := extractTimestampAndMsg(scanner.Text())
			dst <- types.Message{
				Timestamp: ts,
				Text:      msg,
				Instance:  pod,
				Name:      container,
				Namespace: namespace,
			}
		}
		if err := scanner.Err(); err != nil {
			done <- err
			return
		}
	}()

	select {
	case <-ctx.Done():
		logrus.Debug("get-log context cancelled")
		return ctx.Err()
	case err := <-done:
		if err != io.EOF {
			logrus.Debugf("failed to read from pod log: %v", err)
			return err
		}
		return nil
	}
}

// startFunctionPodInformer will gather the list of existing Pods for the function, it will
// watch for newly added or deleted function instances.
func startFunctionPodInformer(ctx context.Context, client kubernetes.Interface, functionName, namespace string) (<-chan string, error) {
	functionSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{consts.LabelInferenceName: functionName},
	}
	selector, err := metav1.LabelSelectorAsSelector(functionSelector)
	if err != nil {
		return nil, errdefs.InvalidParameter(err)
	}

	logrus.WithFields(logrus.Fields{
		"selector":  selector.String(),
		"namespace": namespace,
	}).Debugf("starting log pod informer")
	factory := informers.NewFilteredSharedInformerFactory(
		client,
		podInformerResync,
		namespace,
		withLabels(selector.String()),
	)

	podInformer := factory.Core().V1().Pods()
	podsResp, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errdefs.NotFound(err)
		} else {
			return nil, errdefs.System(err)
		}
	}

	pods := podsResp.Items
	if len(pods) == 0 {
		return nil, errdefs.NotFound(
			fmt.Errorf("no pods found for inference: %s", functionName))
	}

	// prepare channel with enough space for the current instance set
	added := make(chan string, len(pods))
	podInformer.Informer().AddEventHandler(&podLoggerEventHandler{
		added: added,
	})

	// will add existing pods to the chan and then listen for any new pods
	go podInformer.Informer().Run(ctx.Done())
	go func() {
		<-ctx.Done()
		close(added)
	}()

	return added, nil
}

// parseSince returns the time.Duration of the requested Since value _or_ 5 minutes
func parseSince(r *time.Time) *int64 {
	var since int64
	if r == nil || r.IsZero() {
		since = int64(defaultLogSince.Seconds())
		return &since
	}
	since = int64(time.Since(*r).Seconds())
	return &since
}

func extractTimestampAndMsg(logText string) (string, time.Time) {
	// first 32 characters is the k8s timestamp
	parts := strings.SplitN(logText, " ", 2)
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		logrus.WithField("logText", logText).
			Errorf("error parsing timestamp: %s", err)
		return "", time.Time{}
	}

	if len(parts) == 2 {
		return parts[1], ts
	}

	return "", ts
}

func withLabels(selector string) internalinterfaces.TweakListOptionsFunc {
	return func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector
	}
}

type podLoggerEventHandler struct {
	cache.ResourceEventHandler
	added   chan<- string
	deleted chan<- string
}

func (h *podLoggerEventHandler) OnAdd(obj interface{}, isInitialList bool) {
	pod := obj.(*corev1.Pod)
	logrus.WithField("pod", pod.Name).Debugf("log pod informer added a pod")
	h.added <- pod.Name
}

func (h *podLoggerEventHandler) OnUpdate(oldObj, newObj interface{}) {
	// purposefully empty, we don't need to do anything for logs on update
}

func (h *podLoggerEventHandler) OnDelete(obj interface{}) {
	// this may not be needed, the log stream Reader _should_ close on its own without
	// us needing to watch and close it
	// pod := obj.(*corev1.Pod)
	// h.deleted <- pod.Name
}
