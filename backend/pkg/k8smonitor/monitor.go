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
	"fmt"
	"k8s.io/client-go/kubernetes"
)

type Monitor struct {
	nodeMonitor    *NodeMonitor
	serviceMonitor *ServiceMonitor
}

func CreateMonitor(clientset kubernetes.Interface) (*Monitor, error) {
	nodeMonitor, err := CreateNodeMonitor(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to create a node monitor: %v", err)
	}
	serviceMonitor, err := CreateServiceMonitor(clientset)
	if err != nil {
		return nil, fmt.Errorf("failed to create a service monitor: %v", err)
	}
	return &Monitor{
		nodeMonitor:    nodeMonitor,
		serviceMonitor: serviceMonitor,
	}, nil
}

func (m *Monitor) Start() {
	m.nodeMonitor.Start()
	m.serviceMonitor.Start()
}

func (m *Monitor) Stop() {
	m.nodeMonitor.Stop()
	m.serviceMonitor.Stop()
}

func (m *Monitor) IsInternalCIDR(ip string) bool {
	if m == nil {
		return true
	}
	if ip == "" {
		// workaround for now, assuming the telemetry is coming from gateway
		return true
	}
	if m.nodeMonitor.IsPodCIDR(ip) {
		return true
	}
	if m.serviceMonitor.IsServiceIP(ip) {
		return true
	}

	return false
}
