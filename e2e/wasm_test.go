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

package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apiclarity/apiclarity/e2e/utils"

	"gotest.tools/assert"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"

	"github.com/apiclarity/apiclarity/api/client/client/operations"
	"github.com/apiclarity/apiclarity/api/client/models"
)

var wantGetAPIInventoryOKBody = &operations.GetAPIInventoryOKBody{
	Items: []*models.APIInfo{
		{
			HasProvidedSpec:      utils.BoolPtr(false),
			HasReconstructedSpec: utils.BoolPtr(false),
			ID:                   1,
			Name:                 "httpbin.test",
			Port:                 80,
		},
	},
	Total: utils.Int64Ptr(1),
}

func TestWasm(t *testing.T) {
	stopCh := make(chan struct{})
	//defer func() {
	//	stopCh <- struct{}{}
	//	time.Sleep(2 * time.Second)
	//}()
	assert.NilError(t, setupWasmTestEnv(stopCh))

	// wait for httpbin and curl to be restarted after deployment patch. better way of doing it?
	time.Sleep(60*time.Second)

	println("making telemetry from curl to httpbin...")
	assert.NilError(t, utils.HttpReqFromCurlToHttpbin())

	// wait for database to be updated
	time.Sleep(2*time.Second)


	f1 := features.New("telemetry event").
		WithLabel("type", "event").
		Assess("telemetry event exist in DB", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			utils.AssertGetAPIEvents(t, apiclarityAPI, nil)
			return ctx
		}).Feature()

	f2 := features.New("spec").
		WithLabel("type", "spec").
		Assess("spec exist in DB", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			utils.AssertGetAPIInventory(t, apiclarityAPI, wantGetAPIInventoryOKBody)
			return ctx
		}).Feature()

	// test features
	testenv.Test(t, f1, f2)
}

func setupWasmTestEnv(stopCh chan struct{}) error {
	println("Set up wasm test env...")

	helmManager := helm.New(KubeconfigFile)
	println("installing istio...")
	if err := installIstio(helmManager); err != nil {
		return fmt.Errorf("failed to install istio: %v", err)
	}

	println("creating namespace test with istio injection enabled...")
	if err := utils.CreateNamespace(k8sClient, "test", utils.IstioInjectionLabel); err != nil {
		return fmt.Errorf("failed to create test namespace: %v", err)
	}

	println("deploying curl and httpbin to test namespace...")
	if err := utils.InstallHttpbin(helmManager); err != nil {
		return fmt.Errorf("failed to install htpbin: %v", err)
	}
	if err := utils.InstallCurl(); err != nil {
		return fmt.Errorf("failed to install curl: %v", err)
	}

	println("deploying apiclarity with wasm enabled...")
	// helm install --set 'trafficSource.envoyWasm.enabled=true' --set 'trafficSource.envoyWasm.namespaces={test}' --create-namespace apiclarity ../charts/apiclarity -n apiclarity --wait
	if err := utils.InstallAPIClarity(helmManager, "--create-namespace --set 'trafficSource.envoyWasm.namespaces={test}' --set 'trafficSource.envoyWasm.enabled=true --wait'"); err != nil {
		return fmt.Errorf("failed to install apiclarity: %v", err)
	}

	println("waiting for apiclarity to run...")
	if err := utils.WaitForAPIClarityPodRunning(k8sClient); err != nil {
		utils.DescribeAPIClarityDeployment()
		utils.DescribeAPIClarityPods()
		return fmt.Errorf("failed to wait for apiclarity pod to be running: %v", err)
	}

	println("port-forward to apiclarity...")
	utils.PortForwardToAPIClarity(stopCh)

	return nil
}

func installIstio(manager *helm.Manager) error {
	// helm repo add --force-update istio https://istio-release.storage.googleapis.com/charts
	err := manager.RunRepo(helm.WithArgs("add", "--force-update", "istio", "https://istio-release.storage.googleapis.com/charts"))
	if err != nil {
		return fmt.Errorf("failed to run helm repo add --force-update istio https://istio-release.storage.googleapis.com/charts: %v", err)
	}
	// helm repo update
	err = manager.RunRepo(helm.WithArgs("update"))
	if err != nil {
		return fmt.Errorf("failed to run helm repo update: %v", err)
	}
	// helm install istio-base istio/base -n istio-system --create-namespace
	err = manager.RunInstall(helm.WithName("istio-base"), helm.WithChart("istio/base"),
		helm.WithNamespace(utils.IstioNamespace), helm.WithArgs("--create-namespace"))
	if err != nil {
		return fmt.Errorf("failed to run helm install istio-base istio/base -n istio-system --create-namespace: %v", err)
	}
	// helm install istiod istio/istiod -n istio-system --wait
	err = manager.RunInstall(helm.WithName("istiod"), helm.WithChart("istio/istiod"),
		helm.WithNamespace(utils.IstioNamespace), helm.WithArgs("--wait"))
	if err != nil {
		return fmt.Errorf("failed to run helm install istiod istio/istiod -n istio-system --wait: %v", err)
	}
	return nil
}
