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

package tools

import (
	"context"

	uuid "github.com/satori/go.uuid"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	ImagePullSecretNamePrefix = "imagepullsecret"
	SecretType                = "Opaque"
	SecretKey                 = "FuzzerAuthData" // nolint:gosec
)

type ImagePullSecret struct {
	body      string
	name      string
	namespace string
	key       string
}

func (s *ImagePullSecret) Name() string {
	return s.name
}

func (s *ImagePullSecret) Key() string {
	return s.key
}

func (s *ImagePullSecret) Set(value string) {
	s.body = value
}

func (s *ImagePullSecret) Save(_ context.Context, client kubernetes.Interface) error {
	secretDataMap := make(map[string][]byte)
	secretDataMap[s.key] = []byte(s.body)

	newSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: s.name,
		},
		Data: secretDataMap,
		Type: SecretType,
	}

	_, err := client.CoreV1().Secrets(s.namespace).Create(context.TODO(), &newSecret, metav1.CreateOptions{})
	if err != nil {
		return err //nolint:wrapcheck // really want to return the error only
	}

	return nil
}

func (s *ImagePullSecret) Delete(_ context.Context, client kubernetes.Interface) error {
	err := client.CoreV1().Secrets(s.namespace).Delete(context.TODO(), s.name, metav1.DeleteOptions{})
	if err != nil {
		return err //nolint:wrapcheck // really want to return the error only
	}
	return nil
}

func NewSecret(namespace string) (*ImagePullSecret, error) {
	secret := &ImagePullSecret{
		body:      "",
		namespace: namespace,
		name:      ImagePullSecretNamePrefix + "-" + uuid.NewV4().String(),
		key:       SecretKey,
	}
	return secret, nil
}
