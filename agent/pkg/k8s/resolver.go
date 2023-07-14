package k8s

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"

	"github.com/anthhub/forwarder"
	"github.com/phayes/freeport"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	corelister "k8s.io/client-go/listers/core/v1"

	"github.com/tensorchord/openmodelz/agent/errdefs"
)

type Resolver interface {
	Resolve(namespace, name string) (url.URL, error)
	Close(url url.URL)
}

func NewPortForwardingResolver(path string) Resolver {
	return &PortForwardingResolver{
		configPath: path,
		results:    make(map[int]*forwarder.Result),
	}
}

func NewEndpointResolver(lister corelister.EndpointsLister) Resolver {
	return &EndpointResolver{
		EndpointLister: lister,
	}
}

type PortForwardingResolver struct {
	configPath string
	results    map[int]*forwarder.Result
}

func (e *PortForwardingResolver) Resolve(namespace, name string) (url.URL, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return url.URL{}, err
	}

	options := []*forwarder.Option{
		{
			// the local port for forwarding
			LocalPort: port,
			// the k8s pod port
			RemotePort: 8080,
			// the forwarding service name
			ServiceName: "mdz-" + name,
			// namespace default is "default"
			Namespace: namespace,
		},
	}

	ret, err := forwarder.WithForwarders(context.Background(), options, e.configPath)
	if err != nil {
		return url.URL{}, err
	}
	e.results[port] = ret
	// wait forwarding ready
	// the remote and local ports are listed
	_, err = ret.Ready()
	if err != nil {
		return url.URL{}, err
	}

	// the ports are ready
	res, err := url.Parse("http://localhost:" + strconv.Itoa(port))
	return *res, err
}

func (e *PortForwardingResolver) Close(url url.URL) {
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		panic(err)
	}
	logrus.Infof("close port forwarding %d\n", port)
	if e.results[port] == nil {
		logrus.Infof("port forwarding %d not found\n", port)
		return
	}
	logrus.Infof("pointer: %v", e.results[port])
	e.results[port].Close()
}

type EndpointResolver struct {
	EndpointLister corelister.EndpointsLister
}

func (e EndpointResolver) Resolve(namespace, name string) (url.URL, error) {
	svcName := consts.DefaultServicePrefix + name

	svc, err := e.EndpointLister.Endpoints(namespace).Get(svcName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return url.URL{}, errdefs.NotFound(err)
		}
		return url.URL{}, errdefs.System(err)
	}

	if len(svc.Subsets) == 0 {
		return url.URL{}, errdefs.NotFound(
			fmt.Errorf("no subsets for \"%s.%s\"", svcName, namespace))
	}

	all := len(svc.Subsets[0].Addresses)
	if len(svc.Subsets[0].Addresses) == 0 {
		return url.URL{}, errdefs.NotFound(
			fmt.Errorf("no addresses for \"%s.%s\"", svcName, namespace))
	}

	target := rand.Intn(all)

	serviceIP := svc.Subsets[0].Addresses[target].IP
	servicePort := svc.Subsets[0].Ports[0].Port

	urlStr := fmt.Sprintf("http://%s:%d", serviceIP, servicePort)

	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return url.URL{}, errdefs.System(err)
	}

	return *urlRes, nil
}

func (e EndpointResolver) Close(url.URL) {
	// do nothing
}
