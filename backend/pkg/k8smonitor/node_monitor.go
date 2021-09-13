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

package k8smonitor

import (
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type NodeMonitor struct {
	podCIDRsMap *sync.Map // Hold parsed nodes podCIDR (*net.IPNet)
	clientset   kubernetes.Interface
	stopCh      chan struct{}
}

func CreateNodeMonitor(clientset kubernetes.Interface) (*NodeMonitor, error) {
	stopCh := make(chan struct{})
	var podCIDRsMap sync.Map
	return &NodeMonitor{
		podCIDRsMap: &podCIDRsMap,
		clientset:   clientset,
		stopCh:      stopCh,
	}, nil
}

func (m *NodeMonitor) Start() {
	log.Info("Starting Node monitor")
	watchlist := cache.NewListWatchFromClient(m.clientset.CoreV1().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())
	_, controller := cache.NewInformer(
		watchlist,
		&v1.Node{},
		ResyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    m.addNode,
			DeleteFunc: m.deleteNode,
			UpdateFunc: m.updateNode,
		},
	)

	go controller.Run(m.stopCh)
}

func (m *NodeMonitor) Stop() {
	log.Info("Stopping Node monitor")
	close(m.stopCh)
}

func (m *NodeMonitor) addNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Warnf("Object in not a node. %T", obj)
		return
	}
	log.Tracef("Node added: %+v", node)

	m.addPodCIDR(node.Name, node.Spec.PodCIDR)
}

func (m *NodeMonitor) deleteNode(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		log.Warnf("Object in not a node. %T", obj)
		return
	}
	log.Tracef("Node deleted: %+v", node.Name)

	m.deletePodCIDR(node.Name)
}

func (m *NodeMonitor) updateNode(oldObj, newObj interface{}) {
	oldNode, ok := oldObj.(*v1.Node)
	if !ok {
		log.Warnf("Old object in not a node. %T", oldObj)
		return
	}
	newNode, ok := newObj.(*v1.Node)
	if !ok {
		log.Warnf("New object in not a node. %T", newObj)
		return
	}

	if oldNode.Spec.PodCIDR != newNode.Spec.PodCIDR {
		m.deletePodCIDR(oldNode.Name)
		m.addPodCIDR(newNode.Name, newNode.Spec.PodCIDR)
	}

	log.Tracef("Node updated: oldNode=%+v, newNode=%+v", oldNode, newNode)
}

func (m *NodeMonitor) addPodCIDR(key, cidr string) {
	if cidr == "" {
		return
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		log.Errorf("Failed to parse pod CIDR: %v. %v", cidr, err)
		return
	}

	m.podCIDRsMap.Store(key, CreateIPNet(ipnet))
}

func (m *NodeMonitor) deletePodCIDR(key string) {
	m.podCIDRsMap.Delete(key)
}

// IsPodCIDR returns true if the given pod meeting the following conditions:
// 1. Included in one of the node's pod CIDRs
// 2. Not a broadcast IP (the last IP in the CIDR range)
// 3. Not a network identifier IP (the first IP in the CIDR range).
func (m *NodeMonitor) IsPodCIDR(ipStr string) bool {
	ret := false
	m.podCIDRsMap.Range(func(key, value interface{}) bool {
		ipNet, ok := value.(*IPNet)
		if !ok {
			log.Warnf("value is not *IPNet: %T", value)
			// stop iteration
			return false
		}
		ip := net.ParseIP(ipStr)
		if ipNet.Network.Contains(ip) && !ipNet.BroadcastIP.Equal(ip) && !ipNet.NetworkIdentifierIP.Equal(ip) {
			ret = true
			// stop iteration
			return false
		}
		// continue iteration
		return true
	})

	return ret
}
