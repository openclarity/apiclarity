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
	uuid "github.com/satori/go.uuid"
)

var fuzzerJobTemplate = []byte(`apiVersion: batch/v1
kind: Job
objectmeta:
  name: apiclarity-fuzzer
  namespace: apiclarity
  labels:
    app: apiclarity-fuzzer
spec:
  backoffLimit: 0
  ttlSecondsAfterFinished: 300
  template:
    objectmeta:
      name: apiclarity-fuzzer
      namespace: apiclarity
      labels:
        app: apiclarity-fuzzer
    spec:
      restartpolicy: Never
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

type ConfigMapClient struct {
	hClient            kubernetes.Interface
	configMapName      string
	configMapNamespace string
	currentJob         *batchv1.Job
}

func (l *ConfigMapClient) TriggerFuzzingJob(apiID int64, endpoint string, securityItem string, timeBudget string) error {
	logging.Logf("[Fuzzer][ConfigMapClient] TriggerFuzzingJob(%v, %v, %v, %v):: -->", apiID, endpoint, securityItem, timeBudget)

	// Retrieve the env var slice for dynamic parameters from which we will configure our pod
	envVars := l.getEnvs(apiID, endpoint, securityItem, timeBudget)

	// Create job struct
	fuzzerJob, err := l.createFuzzerJob(envVars)
	if err != nil {
		logging.Logf("[Fuzzer][ConfigMapClient] Failed to create fuzzer job struct: %v", err)
		return fmt.Errorf("failed to get create job struct")
	}

	// Create pod from job
	if _, err := l.Create(fuzzerJob); err != nil {
		logging.Logf("[Fuzzer][ConfigMapClient] Failed to create fuzzer job: %v", err)
		return fmt.Errorf("failed to get create job")
	}

	logging.Logf("[Fuzzer][ConfigMapClient] TriggerFuzzingJob():: <--")
	return nil
}

func (l *ConfigMapClient) StopFuzzingJob(apiID int64, complete bool) error {
	logging.Logf("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): -->", apiID)
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
		logging.Logf("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): failed to stop k8s fuzzer job: %v", apiID, err)
	}
	l.currentJob = nil
	logging.Logf("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): <--", apiID)
	return nil
}

func (l *ConfigMapClient) createFuzzerJob(dynEnvVars []v1.EnvVar) (*batchv1.Job, error) {
	var job batchv1.Job
	var fuzzerTemplate []byte

	if l.configMapName == "" {
		// Use default fuzzer job template from config map.
		fuzzerTemplate = fuzzerJobTemplate
	} else {
		// Get fuzzer job template from config map.
		logging.Debugf("[Fuzzer][ConfigMapClient] Load configmap (%v) from namespace (%v)", l.configMapName, l.configMapNamespace)
		cm, err := l.hClient.CoreV1().ConfigMaps(l.configMapNamespace).Get(context.TODO(), l.configMapName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get fuzzer template config map: %v", err)
		}

		config, ok := cm.Data["config"]
		if !ok {
			return nil, fmt.Errorf("no fuzzer template config in configmap")
		}

		fuzzerTemplate = []byte(config)
	}

	logging.Debugf("[Fuzzer][ConfigMapClient] Using fuzzerTemplate:\n%+v", string(fuzzerTemplate))

	err := yaml.Unmarshal(fuzzerTemplate, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fuzzer template: %v", err)
	}

	// Add dynamic env variables to the job existing ones
	containers := job.Spec.Template.Spec.Containers
	if len(containers) != 1 {
		return nil, fmt.Errorf("must have one and only one container in fuzzer template")
	}
	curEnvVars := job.Spec.Template.Spec.Containers[0].Env
	job.Spec.Template.Spec.Containers[0].Env = append(curEnvVars, dynEnvVars...)

	logging.Debugf("[Fuzzer][ConfigMapClient] Using job:\n%+v", job)

	// Check for job name
	if job.GetName() == "" {
		// Manually set one
		job.ObjectMeta.Name = jobNamePrefix + uuid.NewV4().String()
	}

	return &job, nil
}

func (l *ConfigMapClient) getEnvs(apiID int64, endpoint string, securityItem string, timeBudget string) []v1.EnvVar {
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

func (l *ConfigMapClient) Create(job *batchv1.Job) (*batchv1.Job, error) {
	if job == nil {
		return nil, fmt.Errorf("invalid job: nil")
	}

	var ret *batchv1.Job
	var err error

	namespace := job.GetNamespace()
	if len(namespace) == 0 {
		logging.Logf("[Fuzzer][ConfigMapClient] no namespace found in job template. Use the configmapnamespace in place (%v).", l.configMapNamespace)
		namespace = l.configMapNamespace
	}
	logging.Logf("[Fuzzer][ConfigMapClient] Create new Job in namespace: %v/%v, name=%v", namespace, l.configMapNamespace, job.Name)
	if ret, err = l.hClient.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{}); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return nil, fmt.Errorf("failed to create job: %v", err)
		}
		logging.Logf("[Fuzzer][ConfigMapClient] Job already exists: %v", job.Name)
		return ret, nil
	}
	l.currentJob = ret
	logging.Logf("[Fuzzer][ConfigMapClient] Job was created successfully. name=%v, namespace=%v", job.Name, job.Namespace)
	return ret, nil
}

//nolint: ireturn,nolintlint
func NewConfigMapClient(config *config.Config, accessor core.BackendAccessor) (Client, error) {
	client := &ConfigMapClient{
		hClient:            accessor.K8SClient(),
		configMapName:      config.GetJobTemplateConfigMapName(),
		configMapNamespace: config.GetJobNamespace(),
		currentJob:         nil,
	}
	if client.hClient == nil {
		logging.Logf("[Fuzzer][ConfigMapClient] Create new Kubernetes client accessor.K8SClient()=%v", accessor.K8SClient())
		client.hClient, _ = k8smonitor.CreateK8sClientset()
	}
	if client.hClient == nil {
		return nil, fmt.Errorf("missing accessor kubernetes client")
	}
	return client, nil
}
