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
	"os/exec"
	"testing"
	"time"

	"github.com/apiclarity/apiclarity/e2e/utils"
	"github.com/go-openapi/strfmt"
	"gotest.tools/assert"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"

	"github.com/apiclarity/apiclarity/api/client/client/operations"
	"github.com/apiclarity/apiclarity/api/client/models"
)

var wantBodyApiInventory = &operations.GetAPIInventoryOKBody{
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

	///// debug ////
	cmd := exec.Command("kubectl", "-n", "test", "get", "pods")
	out, err := cmd.CombinedOutput()
	assert.NilError(t, err)
	fmt.Printf("kubectl get pods -n test before sleep:\n %s\n", out)
	time.Sleep(60*time.Second)
	cmd = exec.Command("kubectl", "-n", "test", "get", "pods")
	out, err = cmd.CombinedOutput()
	assert.NilError(t, err)
	fmt.Printf("kubectl get pods -n test after sleep:\n %s\n", out)
	///// debug ////

	println("making telemetry from curl to httpbin...")
	assert.NilError(t, utils.HttpReqFromCurlToHttpbin())

	f1 := features.New("telemetry event").
		WithLabel("type", "event").
		Assess("telemetry event exist in UI", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			startTime, err := time.Parse("2006-01-02T15:04:05.000Z", "2021-04-26T11:35:49.775Z")
			assert.NilError(t, err)
			endTime, err := time.Parse("2006-01-02T15:04:05.000Z", "2030-04-26T11:35:49.775Z")
			assert.NilError(t, err)

			params := operations.NewGetAPIEventsParams().WithPage(0).WithPageSize(50).WithStartTime(strfmt.DateTime(startTime)).WithEndTime(strfmt.DateTime(endTime)).WithSortKey("time").WithShowNonAPI(false)
			res, err := apiclarityAPI.Operations.GetAPIEvents(params)
			assert.NilError(t, err)
			assert.Assert(t, *res.Payload.Total == 1)
			// TODO assert payload items...

			return ctx
		}).Feature()

	f2 := features.New("spec").
		WithLabel("type", "spec").
		Assess("spec exist in UI", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			params := operations.NewGetAPIInventoryParams().WithPage(0).WithPageSize(50).WithType(string(models.APITypeINTERNAL)).WithSortKey("name")
			res, err := apiclarityAPI.Operations.GetAPIInventory(params)
			assert.NilError(t, err)
			assert.DeepEqual(t, res.Payload, wantBodyApiInventory)
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
	if err := installHttpbin(helmManager); err != nil {
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

func installHttpbin(manager *helm.Manager) error {
	// helm repo add --force-update matheusfm https://matheusfm.dev/charts
	err := manager.RunRepo(helm.WithArgs("add", "--force-update", "matheusfm", "https://matheusfm.dev/charts"))
	if err != nil {
		return fmt.Errorf("failed to run helm repo add --force-update matheusfm https://matheusfm.dev/charts: %v", err)
	}

	// helm install httpbin matheusfm/httpbin -n test --wait
	err = manager.RunInstall(helm.WithName("httpbin"), helm.WithChart("matheusfm/httpbin"),
		helm.WithNamespace("test"), helm.WithArgs("--wait"))
	if err != nil {
		return fmt.Errorf("failed to run helm install httpbin matheusfm/httpbin  -n test --wait: %v", err)
	}
	return nil
}
