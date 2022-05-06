// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8straceannotator

import (
	"context"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"

	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

func NewK8SAnnotatedK8STelemetry(trace *pluginsmodels.Telemetry, src, dest runtime.Object) *K8SAnnotatedK8STelemetry {
	k8strace := &K8SAnnotatedK8STelemetry{
		RequestID:   trace.RequestID,
		Scheme:      trace.Scheme,
		Destination: &AppEnvInfo{Address: trace.DestinationAddress},
		Source:      &AppEnvInfo{Address: trace.SourceAddress},
		Request:     trace.Request,
		Response:    trace.Response,
	}
	if dest != nil {
		k8strace.Destination.K8SObject = NewRef(dest)
	}
	if src != nil {
		k8strace.Source.K8SObject = NewRef(src)
	}
	return k8strace
}

type K8SAnnotatedK8STelemetry struct {
	RequestID   string
	Scheme      string
	Destination *AppEnvInfo
	Source      *AppEnvInfo
	Request     *pluginsmodels.Request
	Response    *pluginsmodels.Response
}

type AppEnvInfo struct {
	Address   string
	Namespace string
	K8SObject *K8sObjectRef
}

func NewRef(obj runtime.Object) *K8sObjectRef {
	gvk := obj.GetObjectKind().GroupVersionKind()
	metaObj, err := meta.Accessor(obj)
	if err != nil {
		log.Error("k8s object does not implement metadata")
		return nil
	}
	return &K8sObjectRef{
		Kind:       gvk.Kind,
		ApiVersion: gvk.GroupVersion().String(),
		Namespace:  metaObj.GetNamespace(),
		Name:       metaObj.GetName(),
		Uid:        string(metaObj.GetUID()),
	}
}

type K8sClient interface {
	ServicesGet(namespace, name string) (*corev1.Service, error)
	ServicesList(namespace string) ([]*corev1.Service, error)
	PodsList(namespace string) ([]*corev1.Pod, error)

	GetObject(ctx context.Context, apiVersion, kind, namespace, name string) (runtime.Object, error)
	GetObjectOwnerRecursively(ctx context.Context, namespace string, refs []metav1.OwnerReference) (runtime.Object, error)
}

type client struct {
	restMapper      meta.RESTMapper
	informerFactory informers.SharedInformerFactory
	stopCh          <-chan struct{}

	servicesOnce *sync.Once
	podsOnce     *sync.Once

	resourcesOnce map[schema.GroupVersionResource]struct{}
	resourcesMu   *sync.RWMutex
}

const ResyncPeriod = 1 * time.Minute

func NewK8sClient(clientset kubernetes.Interface) (K8sClient, error) {
	stopCh := make(chan struct{})
	rm := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(clientset.Discovery()))
	informerFactory := informers.NewSharedInformerFactory(clientset, ResyncPeriod)
	return &client{
		rm,
		informerFactory,
		stopCh,
		&sync.Once{},
		&sync.Once{},
		map[schema.GroupVersionResource]struct{}{},
		&sync.RWMutex{},
	}, nil
}

func (c *client) GetObjectOwnerRecursively(ctx context.Context, namespace string, refs []metav1.OwnerReference) (runtime.Object, error) {
	for _, ref := range refs {
		obj, err := c.GetObject(ctx, ref.APIVersion, ref.Kind, namespace, ref.Name)
		if err != nil {
			return nil, err
		}
		metaObj, err := meta.Accessor(obj)
		if err != nil {
			return obj, fmt.Errorf("unable to get k8s meta: %w", err)
		}
		ownerObj, err := c.GetObjectOwnerRecursively(ctx, namespace, metaObj.GetOwnerReferences())
		if err != nil {
			return nil, err
		}
		// if no parent is fount return the current object
		if ownerObj == nil {
			return obj, nil
		}

		// nolint:staticcheck
		return ownerObj, nil
	}
	return nil, nil
}

func (c *client) ServicesGet(namespace, name string) (obj *corev1.Service, _ error) {
	informer := c.informerFactory.Core().V1().Services()
	c.servicesOnce.Do(func() {
		go informer.Informer().Run(c.stopCh)
		cache.WaitForCacheSync(c.stopCh, informer.Informer().HasSynced)
		log.Info("synced resources for: Services")
	})
	defer func() {
		_ = addObjectTypeMeta(obj)
	}()
	obj, err := informer.Lister().Services(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("get k8s svc error: %w", err)
	}
	return obj, nil
}

func (c *client) ServicesList(namespace string) (objs []*corev1.Service, err error) {
	informer := c.informerFactory.Core().V1().Services()
	c.servicesOnce.Do(func() {
		go informer.Informer().Run(c.stopCh)
		cache.WaitForCacheSync(c.stopCh, informer.Informer().HasSynced)
		log.Info("synced resources for: Services")
	})
	defer func() {
		for _, i := range objs {
			_ = addObjectTypeMeta(i)
		}
	}()

	objs, err = informer.Lister().Services(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("get k8s svc items error: %w", err)
	}
	return objs, nil
}

func (c *client) PodsList(namespace string) (objs []*corev1.Pod, _ error) {
	informer := c.informerFactory.Core().V1().Pods()
	c.podsOnce.Do(func() {
		go informer.Informer().Run(c.stopCh)
		cache.WaitForCacheSync(c.stopCh, informer.Informer().HasSynced)
		log.Info("synced resources for: Pods")
	})
	defer func() {
		for _, i := range objs {
			_ = addObjectTypeMeta(i)
		}
	}()
	objs, err := informer.Lister().Pods(namespace).List(labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("get k8s pod error: %w", err)
	}
	return objs, nil
}

func (c *client) GetObject(_ context.Context, apiVersion, kind, namespace, name string) (runtime.Object, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to get k8s obj: %w", err)
	}
	gvk := schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    kind,
	}
	mapping, err := c.restMapper.RESTMapping(schema.GroupKind{Kind: kind, Group: gv.Group}, gv.Version)
	if err != nil {
		return nil, fmt.Errorf("cannot map kind/group/version %w", err)
	}
	gvr := schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: mapping.Resource.Resource,
	}
	informer, err := c.informerFactory.ForResource(gvr)
	if err != nil {
		return nil, fmt.Errorf("unable to get k8s informer for gvr=%v: %w", gvr, err)
	}
	c.resourcesMu.Lock()
	_, ok := c.resourcesOnce[gvr]
	if !ok {
		c.resourcesOnce[gvr] = struct{}{}
		go informer.Informer().Run(c.stopCh)
		cache.WaitForCacheSync(c.stopCh, informer.Informer().HasSynced)
		log.Info("synced resources for: ", gvr)
	}
	c.resourcesMu.Unlock()

	_, err = scheme.Scheme.New(gvk)
	if err != nil {
		return nil, fmt.Errorf("unable get k8s obj: %w", err)
	}
	obj, err := informer.Lister().ByNamespace(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve k8s object %w", err)
	}
	obj.GetObjectKind().SetGroupVersionKind(gvk)
	return obj, nil
}

func addObjectTypeMeta(obj runtime.Object) error {
	if !obj.GetObjectKind().GroupVersionKind().Empty() {
		return nil
	}
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}
	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}
	return nil
}
