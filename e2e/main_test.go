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
	"github.com/apiclarity/apiclarity/api/client/client"
	"github.com/apiclarity/apiclarity/e2e/utils"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"os"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"testing"
)

var (
	testenv        env.Environment
	KubeconfigFile string
	apiclarityAPI  *client.APIClarityAPIs
	k8sClient      klient.Client
)

func TestMain(m *testing.M) {
	testenv = env.New()
	kindClusterName := envconf.RandomName("my-cluster", 16)
	//	namespace := envconf.RandomName("myns", 16)

	testenv.Setup(
		envfuncs.CreateKindClusterWithConfig(kindClusterName, "kindest/node:v1.22.2", "kind-config.yaml"),
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error){
			println("Setup")
			k8sClient = cfg.Client()

			tag := os.Getenv("DOCKER_TAG")

			println("DOCKER_TAG=", tag)

			if err := utils.LoadDockerImageToCluster(kindClusterName, fmt.Sprintf("ghcr.io/apiclarity/apiclarity:%v", tag)); err != nil {
				fmt.Printf("Failed to load docker image to cluster: %v", err)
			}
			if err := utils.LoadDockerImageToCluster(kindClusterName, fmt.Sprintf("ghcr.io/apiclarity/kong-plugin:%v", tag)); err != nil {
				fmt.Printf("Failed to load docker image to cluster: %v", err)
			}
			if err := utils.LoadDockerImageToCluster(kindClusterName, fmt.Sprintf("ghcr.io/apiclarity/tyk-plugin-v3.2.2:%v", tag)); err != nil {
				fmt.Printf("Failed to load docker image to cluster: %v", err)
			}
			if err := utils.LoadDockerImageToCluster(kindClusterName, fmt.Sprintf("ghcr.io/apiclarity/passive-taper:%v", tag)); err != nil {
				fmt.Printf("Failed to load docker image to cluster: %v", err)
			}

			clientTransport := httptransport.New("localhost:" + utils.APIClarityPortForwardHostPort, client.DefaultBasePath, []string{"http"})
			apiclarityAPI = client.New(clientTransport, strfmt.Default)

			KubeconfigFile = cfg.KubeconfigFile()

			return ctx, nil
		},
	)

	testenv.Finish(
		func(ctx context.Context, _ *envconf.Config) (context.Context, error){
			println("Finish")
			return ctx, nil
		},
		//envfuncs.DestroyKindCluster(kindClusterName),
	).BeforeEachTest(
		func(ctx context.Context, _ *envconf.Config, _ *testing.T) (context.Context, error){
			println("BeforeEachTest")

			return ctx, nil
		},
	)

	// launch package tests
	os.Exit(testenv.Run(m))
}
