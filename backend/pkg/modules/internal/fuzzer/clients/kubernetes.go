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

package clients

import (
	"context"
	"fmt"
	"strconv"

	uuid "github.com/satori/go.uuid"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/apiclarity/backend/pkg/k8smonitor"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
)

var (
	jobMeta = metav1.TypeMeta{
		Kind:       "Job",
		APIVersion: "batch/v1",
	}

	fuzzerLabels = map[string]string{
		"app": "apiclarity-fuzzer",
	}
)

const (
	convertToBase10 = 10
)

type K8sClient struct {
	hClient                kubernetes.Interface
	namespace              string
	imageName              string
	platformType           string
	platformHostFromFuzzer string
	subFuzzer              string
	tokenInjectorPath      string
	currentJob             *batchv1.Job
}

func (l *K8sClient) TriggerFuzzingJob(apiID int64, endpoint string, securityItem string, timeBudget string) error {
	logging.Logf("[Fuzzer][K8sClient] TriggerFuzzingJob(%v, %v, %v, %v):: -->", apiID, endpoint, securityItem, timeBudget)

	// Retrieve the env var slice that will configure our pod
	envVars := l.getEnvs(apiID, endpoint, securityItem, timeBudget)
	logging.Logf("[Fuzzer][K8sClient] envVars=%v", envVars)

	// Create job struct
	fuzzerJob := l.createFuzzerJob(envVars)

	// Create job item
	if _, err := l.Create(fuzzerJob); err != nil {
		logging.Logf("[Fuzzer][K8sClient] Failed to create fuzzer job: %v", err)
		return fmt.Errorf("failed to get create job: %v", err)
	}

	logging.Logf("[Fuzzer][K8sClient] TriggerFuzzingJob():: <--")
	return nil
}

func (l *K8sClient) StopFuzzingJob(apiID int64, complete bool) error {
	logging.Logf("[Fuzzer][K8sClient] StopFuzzingJob(%v): -->", apiID)
	if l.currentJob == nil {
		return fmt.Errorf("no current k8s job to terminate")
	}
	var zero int64 // = 0
	policy := metav1.DeletePropagationForeground
	deleteOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: &zero,
		PropagationPolicy:  &policy,
	}
	err := l.hClient.BatchV1().Jobs(l.currentJob.Namespace).Delete(context.TODO(), l.currentJob.Name, *deleteOptions)
	if err != nil {
		logging.Logf("[Fuzzer][K8sClient] StopFuzzingJob(%v): failed to stop k8s fuzzer job: %v", apiID, err)
	}
	l.currentJob = nil
	logging.Logf("[Fuzzer][K8sClient] StopFuzzingJob(%v): <--", apiID)
	return nil
}

func (l *K8sClient) createFuzzerJob(envVars []v1.EnvVar) *batchv1.Job {
	var ttlSecondsAfterFinished int32 = 300
	var backOffLimit int32

	containerSecurityContext := CreateSCNJobSecurityContext()

	return &batchv1.Job{
		TypeMeta: jobMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobNamePrefix + uuid.NewV4().String(),
			Namespace: l.namespace,
			Labels:    fuzzerLabels,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fuzzerLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            fuzzerContainerName,
							Image:           l.imageName,
							Env:             envVars,
							SecurityContext: containerSecurityContext,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      tmpEmptyDirVolumeName,
									MountPath: tmpFolderPath,
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyNever,
					SecurityContext: &v1.PodSecurityContext{
						FSGroup: containerSecurityContext.RunAsUser,
					},
					Volumes: []v1.Volume{
						{
							Name: tmpEmptyDirVolumeName,
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			BackoffLimit:            &backOffLimit,
			TTLSecondsAfterFinished: &ttlSecondsAfterFinished,
		},
	}
}

func CreateSCNJobSecurityContext() *v1.SecurityContext {
	trueVal := true
	falseVal := false
	user := int64(user)

	return &v1.SecurityContext{
		Capabilities:             &v1.Capabilities{Drop: []v1.Capability{"ALL"}},
		Privileged:               &falseVal,
		RunAsUser:                &user,
		RunAsGroup:               &user,
		RunAsNonRoot:             &trueVal,
		ReadOnlyRootFilesystem:   &falseVal,
		AllowPrivilegeEscalation: &falseVal,
	}
}

func (l *K8sClient) getEnvs(apiID int64, endpoint string, securityItem string, timeBudget string) []v1.EnvVar {
	envs := []v1.EnvVar{
		{
			Name:  uriEnvVar,
			Value: endpoint,
		},
		{
			Name:  platformTypeEnvVar,
			Value: l.platformType,
		},
		{
			Name:  platformHostEnvVar,
			Value: l.platformHostFromFuzzer,
		},
		{
			Name:  apiIDEnvVar,
			Value: strconv.FormatInt(apiID, convertToBase10),
		},
		{
			Name:  fuzzersEnvVar,
			Value: l.subFuzzer,
		},
		{
			Name:  requestScopeEnvVar,
			Value: requestScopeDefaultValue,
		},
		{
			Name:  restlerRootPathEnvVar,
			Value: tmpFolderPath,
		},
		{
			Name:  authInjectorPathEnvVar,
			Value: l.tokenInjectorPath,
		},
		{
			Name:  restlerTimeBudgetEnvVar,
			Value: timeBudget,
		},
		{
			Name:  debugEnvVar,
			Value: "true",
		},
	}
	if len(securityItem) > 0 {
		envs = append(envs, v1.EnvVar{
			Name:  authEnvVar,
			Value: securityItem,
		})
	}
	return envs
}

func (l *K8sClient) Create(job *batchv1.Job) (*batchv1.Job, error) {
	if job == nil {
		return nil, fmt.Errorf("invalid job: nil")
	}

	var ret *batchv1.Job
	var err error

	logging.Logf("[Fuzzer][K8sClient] Create new Job in namespace: %v/%v, name=%v", job.GetNamespace(), job.Namespace, job.Name)
	if ret, err = l.hClient.BatchV1().Jobs(job.GetNamespace()).Create(context.TODO(), job, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create job: %v", err)
		}
		logging.Logf("[Fuzzer][K8sClient] Job already exists: %v", job.Name)
		return ret, nil
	}
	l.currentJob = ret
	logging.Logf("[Fuzzer][K8sClient] Job was created successfully. name=%v, namespace=%v", job.Name, job.Namespace)
	return ret, nil
}

//nolint: ireturn,nolintlint
func NewKubernetesClient(config *config.Config, accessor core.BackendAccessor) (Client, error) {
	client := &K8sClient{
		hClient:                accessor.K8SClient(),
		imageName:              config.GetImageName(),
		namespace:              config.GetJobNamespace(),
		platformType:           config.GetPlatformType(),
		platformHostFromFuzzer: config.GetPlatformHostFromFuzzer(),
		subFuzzer:              config.GetSubFuzzerList(),
		tokenInjectorPath:      config.GetRestlerTokenInjectorPath(),
		currentJob:             nil,
	}
	if client.hClient == nil {
		logging.Logf("[Fuzzer][K8sClient] Create new Kubernetes client accessor.K8SClient()=%v", accessor.K8SClient())
		client.hClient, _ = k8smonitor.CreateK8sClientset()
	}
	if client.hClient == nil {
		return nil, fmt.Errorf("missing accessor kubernetes client")
	}
	return client, nil
}
