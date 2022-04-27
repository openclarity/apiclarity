package e2e

import (
	"context"
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
		envfuncs.DestroyKindCluster(kindClusterName),
	).BeforeEachTest(
		func(ctx context.Context, _ *envconf.Config, _ *testing.T) (context.Context, error){
			println("BeforeEachTest")

			return ctx, nil
		},
	)

	// launch package tests
	os.Exit(testenv.Run(m))
}
