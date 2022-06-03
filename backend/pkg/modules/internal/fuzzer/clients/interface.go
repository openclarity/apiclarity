package clients

import (
	"errors"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
)

type Client interface {
	TriggerFuzzingJob(apiID uint32, endpoint string, securityItem string) error
}

//nolint: ireturn
func NewClient(moduleConfig *config.Config, accessor core.BackendAccessor) (Client, error) {
	if moduleConfig.GetDeploymentType() == config.DeploymentTypeKubernetes {
		client, err := NewKubernetesClient(moduleConfig, accessor)
		if err != nil {
			logging.Logf("[Fuzzer] Error, can't create Kubernetes client, err=(%v)", err)
			return nil, err
		}
		logging.Logf("[Fuzzer] Docker client creation, ok")
		return client, nil
	} else if moduleConfig.GetDeploymentType() == config.DeploymentTypeDocker {
		client, err := NewDockerClient(moduleConfig)
		if err != nil {
			logging.Logf("[Fuzzer] Error, can't create Docker client, err=(%v)", err)
			return nil, err
		}
		logging.Logf("[Fuzzer] Docker client creation, ok")
		return client, nil
	} else if moduleConfig.GetDeploymentType() == config.DeploymentTypeFake {
		client, err := NewFakeClient(moduleConfig)
		if err != nil {
			logging.Logf("[Fuzzer] Error, can't create Fake client, err=(%v)", err)
			return nil, err
		}
		logging.Logf("[Fuzzer] Docker Fake creation, ok")
		return client, nil
	}

	// ... Else, not supported
	logging.Errorf("[Fuzzer] unsupported DEPLOYMENT_TYPE = (%v)", moduleConfig.GetDeploymentType())
	return nil, errors.New("not supported")
}
