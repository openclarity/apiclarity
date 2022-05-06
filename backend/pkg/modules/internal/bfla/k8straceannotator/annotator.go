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
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

type K8sObjectRef struct {
	// nolint:revive,stylecheck
	ApiVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	// nolint:revive,stylecheck
	Uid string `json:"uid,omitempty"`
}

func DetectSourceObject(ctx context.Context, k8s K8sClient, trace *pluginsmodels.Telemetry) (runtime.Object, error) {
	srcIP, _ := ParseAddr(trace.SourceAddress)
	if srcIP == "" {
		return nil, nil
	}
	pod, err := lookupPods(ctx, k8s, srcIP)
	if err != nil {
		return nil, fmt.Errorf("unable to detect src object: %w", err)
	}
	if pod == nil {
		return nil, fmt.Errorf("unable to find source k8s object for ip: %s", srcIP)
	}
	obj, err := k8s.GetObjectOwnerRecursively(ctx, pod.Namespace, pod.GetOwnerReferences())
	if err != nil {
		return nil, fmt.Errorf("unable to detect src object recursevly: %w", err)
	}
	if obj != nil {
		return obj, nil
	}
	return pod, nil
}

func DetectDestinationObject(ctx context.Context, k8s K8sClient, trace *pluginsmodels.Telemetry) (runtime.Object, error) {
	destIP, _ := ParseAddr(trace.DestinationAddress)
	if destIP != "" {
		svc, err := lookupServices(ctx, k8s, destIP)
		if err != nil {
			return nil, fmt.Errorf("unable to detect dest object: %w", err)
		}
		if svc != nil {
			return svc, nil
		}

		// source ip (Pod -> ReplicaSet -> Deployment) => destination ip (Service -x Deployment).
		pod, err := lookupPods(ctx, k8s, destIP)
		if err != nil {
			return nil, fmt.Errorf("unable to detect dest object: %w", err)
		}
		if pod != nil {
			return pod, nil
		}

		return nil, fmt.Errorf("unable to find destination k8s object")
	}
	if trace.Request.Host == "" {
		return nil, nil
	}
	svcAndNs := strings.Split(trace.Request.Host, ".")
	// nolint:gomnd
	if len(svcAndNs) == 2 {
		svc, err := k8s.ServicesGet(svcAndNs[1], svcAndNs[0])
		if err != nil {
			return nil, fmt.Errorf("unable to detect dest object: %w", err)
		}
		return svc, nil
	}

	return nil, fmt.Errorf("unexpected host name: %s", trace.Request.Host)
}

func lookupServices(_ context.Context, k8s K8sClient, wantIP string) (*corev1.Service, error) {
	services, err := k8s.ServicesList("")
	if err != nil {
		return nil, fmt.Errorf("unable to lookup k8s services: %w", err)
	}
	for _, svc := range services {
		for _, ip := range svc.Spec.ClusterIPs {
			if ip == wantIP {
				return svc, nil
			}
		}
	}
	return nil, nil
}

func lookupPods(_ context.Context, k8s K8sClient, wantIP string) (*corev1.Pod, error) {
	pods, err := k8s.PodsList("")
	if err != nil {
		return nil, fmt.Errorf("unable to lookup pods: %w", err)
	}
	for _, pod := range pods {
		for _, ip := range pod.Status.PodIPs {
			if ip.IP == wantIP {
				return pod, nil
			}
		}
	}
	return nil, nil
}

func ParseAddr(addr string) (ip, port string) {
	parts := strings.Split(addr, ":")
	// nolint:gomnd
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return addr, ""
}
