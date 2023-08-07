// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"fmt"
	"math/rand"
	"net/url"
	"strings"
	"sync"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	corelister "k8s.io/client-go/listers/core/v1"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

func NewFunctionLookup(ns string, lister corelister.EndpointsLister) *FunctionLookup {
	return &FunctionLookup{
		DefaultNamespace: ns,
		EndpointLister:   lister,
		Listers:          map[string]corelister.EndpointsNamespaceLister{},
		lock:             sync.RWMutex{},
	}
}

type FunctionLookup struct {
	DefaultNamespace string
	EndpointLister   corelister.EndpointsLister
	Listers          map[string]corelister.EndpointsNamespaceLister

	lock sync.RWMutex
}

func (f *FunctionLookup) GetLister(ns string) corelister.EndpointsNamespaceLister {
	f.lock.RLock()
	defer f.lock.RUnlock()
	return f.Listers[ns]
}

func (f *FunctionLookup) SetLister(ns string, lister corelister.EndpointsNamespaceLister) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.Listers[ns] = lister
}

func getNamespace(name, defaultNamespace string) string {
	namespace := defaultNamespace
	if strings.Contains(name, ".") {
		namespace = name[strings.LastIndexAny(name, ".")+1:]
	}
	return namespace
}

func (l *FunctionLookup) Resolve(name string) (url.URL, error) {
	functionName := name
	namespace := getNamespace(name, l.DefaultNamespace)
	if err := l.verifyNamespace(namespace); err != nil {
		return url.URL{}, err
	}

	if strings.Contains(name, ".") {
		functionName = strings.TrimSuffix(name, "."+namespace)
	}

	nsEndpointLister := l.GetLister(namespace)

	if nsEndpointLister == nil {
		l.SetLister(namespace, l.EndpointLister.Endpoints(namespace))

		nsEndpointLister = l.GetLister(namespace)
	}

	svcName := consts.DefaultServicePrefix + functionName

	svc, err := nsEndpointLister.Get(svcName)
	if err != nil {
		return url.URL{}, fmt.Errorf("error listing \"%s.%s\": %s", svcName, namespace, err.Error())
	}

	if len(svc.Subsets) == 0 {
		return url.URL{}, fmt.Errorf("no subsets available for \"%s.%s\"", svcName, namespace)
	}

	all := len(svc.Subsets[0].Addresses)
	if len(svc.Subsets[0].Addresses) == 0 {
		return url.URL{}, fmt.Errorf("no addresses in subset for \"%s.%s\"", svcName, namespace)
	}

	target := rand.Intn(all)

	serviceIP := svc.Subsets[0].Addresses[target].IP
	servicePort := svc.Subsets[0].Ports[target].Port

	urlStr := fmt.Sprintf("http://%s:%d", serviceIP, servicePort)

	urlRes, err := url.Parse(urlStr)
	if err != nil {
		return url.URL{}, err
	}

	return *urlRes, nil
}

func (l *FunctionLookup) verifyNamespace(name string) error {
	if name != "kube-system" {
		return nil
	}
	// ToDo use global namespace parse and validation
	return fmt.Errorf("namespace not allowed")
}
