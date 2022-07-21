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
	"strings"

	"gopkg.in/yaml.v2"
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

var fuzzerJobTemplate = []byte(`apiVersion: batch/v1
kind: Job
metadata:
  name: apiclarity-fuzzer
  namespace: apiclarity
  labels:
    app: apiclarity-fuzzer
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 300
  template:
    metadata:
      name: apiclarity-fuzzer
      namespace: apiclarity
      labels:
        app: apiclarity-fuzzer
    spec:
      restartPolicy: Never
      containers:
      - name: fuzzer
        volumeMounts:
        - name: tmp-volume
          mountPath: /tmp
      image: gcr.io/eticloud/k8sec/scn-dast:9a2af104203225dd39ed07f279d9f4cdc1053aa3 
      env:
      - name: PLATFORM_TYPE
        value: "API_CLARITY"
      - name: PLATFORM_HOST
        value: "http://apiclarity-apiclarity:8080/api"
      - name: FUZZER
        value: "scn-fuzzer,restler,crud"
      - name: REQUEST_SCOPE
        value: "global/internalservices/portshift_request"
      - name: RESTLER_ROOT_PATH
        value: "/tmp"
      - name: RESTLER_TOKEN_INJECTOR_PATH
        value: "/app/"
      - name: DEBUG
        value: true
      - name: URI
        value: <URI>
      - name: API_ID
        value: <API_ID>
      - name: RESTLER_TIME_BUDGET
        value: <RESTLER_TIME_BUDGET>
      securityContext:
        capabilities:
          drop:
          - all
        runAsNonRoot: true
        runAsGroup: 1001
        runAsUser: 1001
        privileged: false
        allowPrivilegeEscalation: false
      readOnlyRootFilesystem: false
      resources:
      requests:
        memory: "50Mi"
        cpu: "50m"
      limits:
        memory: "1000Mi"
        cpu: "1000m"
      volumes:
        - name: tmp-volume
          emptyDir: {}
`)

const ()

type HelmChartClient struct {
	hClient            kubernetes.Interface
	configMapName      string
	configMapNamespace string
	currentJob         *batchv1.Job
}

func (l *HelmChartClient) TriggerFuzzingJob(apiID int64, endpoint string, securityItem string, timeBudget string) error {
	logging.Logf("[Fuzzer][HelmChartClient] TriggerFuzzingJob(%v, %v, %v, %v):: -->", apiID, endpoint, securityItem, timeBudget)

	// Retrieve the env var slice that will configure our pod
	envVars := l.getEnvs(apiID, endpoint, securityItem, timeBudget)
	logging.Logf("[Fuzzer][HelmChartClient] envVars=%v", envVars)

	// Create job struct
	fuzzerJob, err := l.createFuzzerJob(envVars)
	if err != nil {
		logging.Logf("[Fuzzer][HelmChartClient] Failed to create fuzzer job struct: %v", err)
		return fmt.Errorf("failed to get create job struct")
	}

	// Create job
	if _, err := l.Create(fuzzerJob); err != nil {
		logging.Logf("[Fuzzer][HelmChartClient] Failed to create fuzzer job: %v", err)
		return fmt.Errorf("failed to get create job")
	}

	logging.Logf("[Fuzzer][HelmChartClient] TriggerFuzzingJob():: <--")
	return nil
}

func (l *HelmChartClient) StopFuzzingJob(apiID int64, complete bool) error {
	logging.Logf("[Fuzzer][HelmChartClient] StopFuzzingJob(%v): -->", apiID)
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
		logging.Logf("[Fuzzer][HelmChartClient] StopFuzzingJob(%v): failed to stop k8s fuzzer job: %v", apiID, err)
	}
	l.currentJob = nil
	logging.Logf("[Fuzzer][HelmChartClient] StopFuzzingJob(%v): <--", apiID)
	return nil
}

func (l *HelmChartClient) createFuzzerJob(envVars []v1.EnvVar) (*batchv1.Job, error) {
	var job batchv1.Job
	var fuzzerTemplate []byte

	if l.configMapName == "" {
		// Use default scanner job template from config map.
		fuzzerTemplate = fuzzerJobTemplate
	} else {
		// Get scanner job template from config map.
		cm, err := l.hClient.CoreV1().ConfigMaps(l.configMapNamespace).Get(context.TODO(), l.configMapName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get scanner template config map: %v", err)
		}

		config, ok := cm.Data["config"]
		if !ok {
			return nil, fmt.Errorf("no scanner template config in configmap")
		}

		fuzzerTemplate = []byte(config)
	}

	fuzzerTemplateStr := string(fuzzerTemplate)
	for _, item := range envVars {
		fuzzerTemplateStr = strings.Replace(fuzzerTemplateStr, "<"+item.Name+">", item.Value, 1)
	}
	fuzzerTemplate = []byte(fuzzerTemplateStr)

	logging.Logf("Using fuzzerTemplate:\n%+v", string(fuzzerTemplate))

	err := yaml.Unmarshal(fuzzerTemplate, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal scanner template: %v", err)
	}

	/*return &batchv1.Job{
		TypeMeta: jobMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name: jobNamePrefix + uuid.NewV4().String(),
			//Namespace: l.namespace,
			Labels: fuzzerLabels,
		},
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: fuzzerLabels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: fuzzerContainerName,
							//Image:           l.imageName,
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
	}*/

	return &job, nil
}

func (l *HelmChartClient) getEnvs(apiID int64, endpoint string, securityItem string, timeBudget string) []v1.EnvVar {
	envs := []v1.EnvVar{
		{
			Name:  uriEnvVar,
			Value: endpoint,
		},
		{
			Name:  apiIDEnvVar,
			Value: strconv.FormatInt(apiID, convertToBase10),
		},
		{
			Name:  restlerTimeBudgetEnvVar,
			Value: timeBudget,
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

func (l *HelmChartClient) Create(job *batchv1.Job) (*batchv1.Job, error) {
	if job == nil {
		return nil, fmt.Errorf("invalid job: nil")
	}

	var ret *batchv1.Job
	var err error

	logging.Logf("[Fuzzer][HelmChartClient] Create new Job in namespace: %v/%v, name=%v", job.GetNamespace(), job.Namespace, job.Name)
	if ret, err = l.hClient.BatchV1().Jobs(job.GetNamespace()).Create(context.TODO(), job, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create job: %v", err)
		}
		logging.Logf("[Fuzzer][HelmChartClient] Job already exists: %v", job.Name)
		return ret, nil
	}
	l.currentJob = ret
	logging.Logf("[Fuzzer][HelmChartClient] Job was created successfully. name=%v, namespace=%v", job.Name, job.Namespace)
	return ret, nil
}

//nolint: ireturn,nolintlint
func NewHelmChartClient(config *config.Config, accessor core.BackendAccessor) (Client, error) {
	client := &HelmChartClient{
		hClient:    accessor.K8SClient(),
		currentJob: nil,
	}
	if client.hClient == nil {
		logging.Logf("[Fuzzer][HelmChartClient] Create new Kubernetes client accessor.K8SClient()=%v", accessor.K8SClient())
		client.hClient, _ = k8smonitor.CreateK8sClientset()
	}
	if client.hClient == nil {
		return nil, fmt.Errorf("missing accessor kubernetes client")
	}
	return client, nil
}
