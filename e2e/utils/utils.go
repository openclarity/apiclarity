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

package utils

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"path/filepath"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/third_party/helm"
	"time"
)

// EXPORTED:

var IstioInjectionLabel = map[string]string{
	"istio-injection": "enabled",
}

func InstallCurl() error {
	cmd := exec.Command("kubectl", "-n", "test", "apply", "-f", "curl.yaml")

	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func DescribeAPIClarityDeployments() {
	cmd := exec.Command("kubectl", "-n", "apiclarity", "describe", "deployments.apps", APIClarityDeploymentName)

	out, err := cmd.CombinedOutput()
	if err != nil {
		println(err)
	}
	fmt.Printf("kubectl describe deployments.apps -n apiclarity apiclarity-apiclarity:\n %s", out)
}

func HttpReqFromCurlToHttpbin() error {
	cmd := exec.Command("kubectl", "-n", "test", "exec", "-it", fmt.Sprintf("%s/%s", "service", "curl"), "-c", "curl", "--", "curl", "-H", "Content-Type: application/json", "httpbin.test.svc.cluster.local:80/get")

	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

var curDir, _ = os.Getwd()
var chartPath = filepath.Join(curDir, "../charts/apiclarity")

func InstallAPIClarity(manager *helm.Manager, args string) error {
	if err := manager.RunInstall(helm.WithName(APIClarityHelmReleaseName),
		helm.WithVersion("v1.1"),
		helm.WithNamespace(APIClarityNamespace),
		helm.WithChart(chartPath),
		helm.WithArgs(args)); err != nil {
		return fmt.Errorf("failed to run helm install command with args: %v. %v", args, err)
	}
	return nil
}

func PortForwardToAPIClarity(stopCh chan struct{}) {
	// TODO make it better
	go func() {
		err, out := portForward("service", APIClarityNamespace, APIClarityServiceName, APIClarityPortForwardHostPort, APIClarityPortForwardTargetPort, stopCh)
		if err != nil {
			println("port forward fail!. %s, %v", out, err)
			return
		}
	}()
	time.Sleep(3 * time.Second)
}

func CreateNamespace(client klient.Client ,name string, labels map[string]string) error {
	var ns = v1.Namespace{
		TypeMeta:   v12.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: v12.ObjectMeta{
			Name:                       name,
			Labels:                     labels,
		},
	}
	if err := client.Resources(name).Create(context.TODO(), &ns); err != nil {
		return err
	}
	return nil
}

func BoolPtr(val bool) *bool {
	ret := val
	return &ret
}

func Int64Ptr(val int64) *int64 {
	ret := val
	return &ret
}

// NON EXPORTED:

func portForward(kind, namespace, name, hostPort, targetPort string, stopCh chan struct{}) (error, []byte) {
	cmd := exec.Command("kubectl", "port-forward", "-n", namespace,
		fmt.Sprintf("%s/%s", kind, name), fmt.Sprintf("%s:%s", hostPort, targetPort))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return err, out
	}
	return nil, nil
}

//TODO use https://github.com/kubernetes-sigs/e2e-framework/tree/main/examples/wait_for_resources
func WaitForAPIClarityPodRunning(client klient.Client) error {
	podList := v1.PodList{}
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	timeout := time.NewTimer(3 * time.Minute)
	for {
		select {
		case <-timeout.C:
			return fmt.Errorf("timeout reached")
		case <-ticker.C:
			if err := client.Resources(APIClarityNamespace).List(context.TODO(), &podList, func(lo *v12.ListOptions){
				lo.LabelSelector = "app=apiclarity-apiclarity"
			}); err != nil {
				return fmt.Errorf("failed to get pod apiclarity. %v", err)
			}
			pod := podList.Items[0]
			if pod.Status.Phase == v1.PodRunning {
				return nil
			}
		}
	}
}
