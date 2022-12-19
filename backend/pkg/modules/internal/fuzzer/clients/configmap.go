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

	logging "github.com/sirupsen/logrus"

	"github.com/ghodss/yaml"
	uuid "github.com/satori/go.uuid"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/apiclarity/backend/pkg/k8smonitor"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/tools"
)

const (
	convertToBase10 = 10
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
      securityContext:
        capabilities:
          drop:
          - all
        runAsNonRoot: true
        runAsGroup: 1001
        runAsUser: 1001
        privileged: false
        allowPrivilegeEscalation: false
        readOnlyRootFilesystem: true
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
	k8sClient          kubernetes.Interface
	configMapName      string
	configMapNamespace string
	jobs               map[int64]*batchv1.Job // List of jobs, per api
	fuzzerJobTemplate  []byte
	authSecrets        map[int64]*tools.AuthSecret // List of secrets, per api
}

func (l *ConfigMapClient) TriggerFuzzingJob(apiID int64, endpoint string, securityItem string, timeBudget string) error {
	logging.Infof("[Fuzzer][ConfigMapClient] TriggerFuzzingJob(%v, %v, %v, %v):: --> <--", apiID, endpoint, securityItem, timeBudget)

	if l.jobs[apiID] != nil {
		err := fmt.Errorf("job already started for the API: %v", apiID)
		logging.Errorf("[Fuzzer][ConfigMapClient] failed to create job, err=(%v)", err)
		return fmt.Errorf("failed to create job, err=(%v)", err)
	}

	// Retrieve the env var slice for dynamic parameters from which we will configure our pod
	envVars := l.getEnvs(apiID, endpoint, securityItem, timeBudget)

	// Create job struct
	jobToCreate, err := l.getFuzzerJobToCreate(envVars)
	if err != nil {
		logging.Errorf("[Fuzzer][ConfigMapClient] Failed to create fuzzer job struct, err=(%v)", err)
		return fmt.Errorf("failed to get create job struct, err=(%v)", err)
	}

	// Create pod from job
	job, err := l.CreateJob(jobToCreate)
	if err != nil {
		logging.Errorf("[Fuzzer][ConfigMapClient] Failed to create fuzzer job, err=(%v)", err)
		return fmt.Errorf("failed to get create job, err=(%v)v", err)
	}
	l.jobs[apiID] = job

	return nil
}

func (l *ConfigMapClient) StopFuzzingJob(apiID int64, complete bool) error {
	logging.Infof("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): --> <--", apiID)

	jobToDelete := l.jobs[apiID]
	if jobToDelete == nil {
		err := fmt.Errorf("no existing fuzzing job to terminate for the API (%v)", apiID)
		return fmt.Errorf("failed to stop job, err=(%v)", err)
	}

	secret, found := l.authSecrets[apiID]
	if found {
		err := secret.Delete(context.TODO(), l.k8sClient)
		if err != nil {
			// Not blocking
			logging.Infof("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): failed to delete secret: %v", apiID, err)
		}
	}

	jobClient := l.k8sClient.BatchV1().Jobs(jobToDelete.Namespace)
	jobName := jobToDelete.Name
	job, err := jobClient.Get(context.TODO(), jobName, metav1.GetOptions{})
	if err != nil {
		// reset job in case of error, otherwise we will be stuck for the API
		l.jobs[apiID] = nil
		return fmt.Errorf("can't find k8s job (%v) to terminate", jobName)
	}
	if job.Status.Active > 0 {
		// the Job is still running, we must stop it
		var zero int64 // = 0
		policy := metav1.DeletePropagationForeground
		deleteOptions := &metav1.DeleteOptions{
			GracePeriodSeconds: &zero,
			PropagationPolicy:  &policy,
		}
		err := jobClient.Delete(context.TODO(), jobName, *deleteOptions)
		if (err != nil) && !k8serrors.IsNotFound(err) {
			// An IsNotFound is not an error, here: just the job ended between the GET and the DELETE. it is Ok for us.
			logging.Infof("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): failed to stop k8s fuzzer job: %v", apiID, err)
		}
	} else {
		// job.Status.Succeeded > 0 mean "Job Successful"
		if job.Status.Succeeded == 0 {
			// The Job has been in error, then nothing to stop.
			logging.Infof("[Fuzzer][ConfigMapClient] StopFuzzingJob(%v): failed to stop k8s fuzzer job: %v", apiID, err)
		}
	}

	l.jobs[apiID] = nil
	return nil
}

func (l *ConfigMapClient) getFuzzerJobToCreate(dynEnvVars []v1.EnvVar) (*batchv1.Job, error) {
	var job batchv1.Job
	logging.Debugf("[Fuzzer][ConfigMapClient] Using fuzzerTemplate:\n%+v", string(l.fuzzerJobTemplate))

	err := yaml.Unmarshal(l.fuzzerJobTemplate, &job)
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
		// create the secret
		secret, err := tools.NewSecret(l.configMapNamespace)
		if err != nil {
			logging.Errorf("Failed to create Secret, err=(%v)", err)
			return envs
		}
		secret.Set(securityItem)
		err = secret.Save(context.TODO(), l.k8sClient)
		if err != nil {
			logging.Errorf("Failed to write the Secret, err=(%v)", err)
			return envs
		}
		l.authSecrets[apiID] = secret

		// pass the secret in Fuzzer pod container env
		envs = append(envs, v1.EnvVar{
			Name: authEnvVar,
			ValueFrom: &v1.EnvVarSource{
				SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{
						Name: secret.Name(),
					},
					Key: secret.Key(),
				},
			},
		})
	}
	return envs
}

func (l *ConfigMapClient) loadConfigMap() ([]byte, error) {
	var fuzzerTemplate []byte

	if l.configMapName == "" {
		// Use default fuzzer job template from config map.
		fuzzerTemplate = fuzzerJobTemplate
	} else {
		// Get fuzzer job template from config map.
		logging.Debugf("[Fuzzer][ConfigMapClient] Load configmap (%v) from namespace (%v)", l.configMapName, l.configMapNamespace)
		cm, err := l.k8sClient.CoreV1().ConfigMaps(l.configMapNamespace).Get(context.TODO(), l.configMapName, metav1.GetOptions{})
		if err != nil {
			return nil, err //nolint:wrapcheck // really want to return the error that come from k8sClient
		}

		config, ok := cm.Data["config"]
		if !ok {
			return nil, fmt.Errorf("no fuzzer template config in configmap")
		}

		fuzzerTemplate = []byte(config)
	}
	return fuzzerTemplate, nil
}

func (l *ConfigMapClient) CreateJob(job *batchv1.Job) (*batchv1.Job, error) {
	if job == nil {
		return nil, fmt.Errorf("invalid job: nil")
	}

	// Define nemaspace to use
	namespace := job.GetNamespace()
	if len(namespace) == 0 {
		logging.Infof("[Fuzzer][ConfigMapClient] no namespace found in job template. Use the configmapnamespace in place (%v).", l.configMapNamespace)
		namespace = l.configMapNamespace
	}

	// Create the k8s job
	logging.Infof("[Fuzzer][ConfigMapClient] Create new Job in namespace: %v/%v, name=%v", namespace, l.configMapNamespace, job.Name)
	newJob, err := l.k8sClient.BatchV1().Jobs(namespace).Create(context.TODO(), job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: (%v)", err)
	}
	logging.Debugf("[Fuzzer][ConfigMapClient] Job was created successfully. name=%v, namespace=%v", job.Name, job.Namespace)
	return newJob, nil
}

// nolint: ireturn,nolintlint
func NewConfigMapClient(config *config.Config, accessor core.BackendAccessor) (Client, error) {
	client := &ConfigMapClient{
		k8sClient:          accessor.K8SClient(),
		configMapName:      config.GetJobTemplateConfigMapName(),
		configMapNamespace: config.GetJobNamespace(),
		jobs:               make(map[int64]*batchv1.Job),
		authSecrets:        make(map[int64]*tools.AuthSecret),
	}
	if client.k8sClient == nil {
		logging.Infof("[Fuzzer][ConfigMapClient] Create new Kubernetes client accessor.K8SClient()=%v", accessor.K8SClient())
		client.k8sClient, _ = k8smonitor.CreateK8sClientset()
	}
	if client.k8sClient == nil {
		return nil, fmt.Errorf("missing accessor kubernetes client")
	}
	fuzzerJobTemplate, err := client.loadConfigMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get fuzzer template config map: %v", err)
	}
	client.fuzzerJobTemplate = fuzzerJobTemplate
	return client, nil
}
