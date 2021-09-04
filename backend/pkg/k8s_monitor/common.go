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

package k8s_monitor

import (
	"encoding/binary"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"time"
)

const ResyncPeriod = 1 * time.Minute

func CreateK8sClientset() (kubernetes.Interface, error) {
	// Create Kubernetes go-client clientset
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %v", err)
	}

	// Create a rest client not targeting specific API version
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create a rest client: %v", err)
	}

	return clientset, nil
}

type IPNet struct {
	Network             *net.IPNet
	NetworkIdentifierIP net.IP
	BroadcastIP         net.IP
}

func CreateIPNet(ipnet *net.IPNet) *IPNet {
	return &IPNet{
		Network:             ipnet,
		NetworkIdentifierIP: getNetworkIdentifierIP(ipnet),
		BroadcastIP:         getBroadcastIP(ipnet),
	}
}

func getBroadcastIP(n *net.IPNet) net.IP {
	networkIP := n.IP.To4()
	if networkIP == nil {
		// IPv6 does not support broadcast addresses
		return nil
	}

	broadcastIP := make(net.IP, len(networkIP))
	binary.BigEndian.PutUint32(broadcastIP, binary.BigEndian.Uint32(networkIP)|^binary.BigEndian.Uint32(net.IP(n.Mask).To4()))

	return broadcastIP
}

func getNetworkIdentifierIP(n *net.IPNet) net.IP {
	ip := n.IP.To4()
	if ip == nil {
		ip = n.IP
		if len(ip) != net.IPv6len {
			return nil
		}
	}

	return ip
}
