// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Portshift/go-utils/k8s"
	log "github.com/sirupsen/logrus"
	"github.com/up9inc/mizu/tap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	ResyncPeriod       = 1 * time.Minute
	Interval           = 5
	KubernetesAPIQPS   = 80.0
	KubernetesAPIBurst = 60
)

type PodMonitor struct {
	clientset       *kubernetes.Clientset
	stopCh          chan struct{}
	podStore        cache.Store
	isStateChange   bool
	namespacesToTap map[string]bool
	lock            sync.RWMutex
}

func NewPodMonitor(namespaces []string) (*PodMonitor, error) {
	clientset, _, err := k8s.CreateK8sClientset(nil, k8s.KubeOptions{
		KubernetesAPIQPS:   KubernetesAPIQPS,
		KubernetesAPIBurst: KubernetesAPIBurst,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create K8s clientset: %v", err)
	}
	return &PodMonitor{
		clientset:       clientset,
		stopCh:          make(chan struct{}),
		isStateChange:   true,
		namespacesToTap: sliceToMap(namespaces),
	}, nil
}

func sliceToMap(s []string) map[string]bool {
	ret := make(map[string]bool)

	for _, s2 := range s {
		ret[s2] = true
	}
	return ret
}

func (pm *PodMonitor) getIsStateChange() bool {
	pm.lock.RLock()
	defer pm.lock.RUnlock()

	return pm.isStateChange
}

func (pm *PodMonitor) setIsStateChange(val bool) {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	pm.isStateChange = val
}

func (pm *PodMonitor) Start(ctx context.Context) {
	watchlist := cache.NewListWatchFromClient(pm.clientset.CoreV1().RESTClient(), v1.ResourcePods.String(), v1.NamespaceAll, fields.Everything())
	store, controller := cache.NewInformer(
		watchlist,
		&v1.Pod{},
		ResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod) //nolint:forcetypeassert
				if pm.namespacesToTap[pod.Namespace] {
					pm.setIsStateChange(true)
				}
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod) //nolint:forcetypeassert
				if pm.namespacesToTap[pod.Namespace] {
					pm.setIsStateChange(true)
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				oldPod := oldObj.(*v1.Pod) //nolint:forcetypeassert
				newPod := newObj.(*v1.Pod) //nolint:forcetypeassert
				if oldPod.Status.PodIP != newPod.Status.PodIP && pm.namespacesToTap[newPod.Namespace] {
					pm.setIsStateChange(true)
				}
			},
		},
	)
	pm.podStore = store

	go controller.Run(pm.stopCh)

	ticker := time.NewTicker(Interval * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Info("Done")
			return
		case <-ticker.C:
			if pm.getIsStateChange() {
				log.Info("Pod ips changed, updating filter authorities")
				pm.updateFilteredAuthorities()
			}
		}
	}
}

func (pm *PodMonitor) updateFilteredAuthorities() {
	pm.lock.Lock()
	defer pm.lock.Unlock()

	podIPs := pm.GetPodsIps()
	tap.SetFilterAuthorities(podIPs)
	pm.isStateChange = false
}

func (pm *PodMonitor) GetPodsIps() []string {
	var podIPs []string

	for _, p := range pm.podStore.List() {
		pod := p.(*v1.Pod) //nolint:forcetypeassert
		if pm.namespacesToTap[pod.Namespace] && pod.Status.PodIP != "" {
			podIPs = append(podIPs, pod.Status.PodIP)
		}
	}
	return podIPs
}

func (pm *PodMonitor) GetPodNamespaceByIP(ip string) string {
	if ip == "" {
		return ""
	}
	for _, p := range pm.podStore.List() {
		pod := p.(*v1.Pod) //nolint:forcetypeassert
		if pod.Status.PodIP == ip {
			return pod.Namespace
		}
	}
	log.Infof("IP %v not found in pod map", ip)
	return ""
}

func (pm *PodMonitor) IsMonitoredNamespace(ns string) bool {
	return pm.namespacesToTap[ns]
}
