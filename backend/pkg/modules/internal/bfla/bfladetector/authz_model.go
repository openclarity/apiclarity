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

package bfladetector

import (
	"bytes"
	"fmt"

	"github.com/apiclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
)

type Operation struct {
	Method   string   `json:"method"`
	Path     string   `json:"path"`
	Audience Audience `json:"audience"`
}

type SourceObject struct {
	K8sObject  *k8straceannotator.K8sObjectRef `json:"k8s_object"`
	External   bool                            `json:"external"`
	EndUsers   EndUsers                        `json:"end_users,omitempty"`
	Authorized bool                            `json:"authorized"`
}

type DetectedUserSource int32

const (
	DetectedUserSourceJWT = iota
	DetectedUserSourceBasic
	DetectedUserSourceXConsumerIDHeader
)

func (d *DetectedUserSource) UnmarshalJSON(b []byte) error {
	buff := bytes.NewBuffer(b)
	srcName := ""
	fmt.Fscanf(buff, "%q", &srcName)
	switch srcName {
	case "JWT":
		*d = DetectedUserSourceJWT
	case "BASIC":
		*d = DetectedUserSourceBasic
	case "KONG_X_CONSUMER_ID":
		*d = DetectedUserSourceXConsumerIDHeader
	}
	return nil
}

func (d DetectedUserSource) MarshalJSON() ([]byte, error) {
	buff := &bytes.Buffer{}
	fmt.Fprintf(buff, "%q", d)
	return buff.Bytes(), nil
}

func (d DetectedUserSource) String() string {
	switch d {
	case DetectedUserSourceJWT:
		return "JWT"
	case DetectedUserSourceBasic:
		return "BASIC"
	case DetectedUserSourceXConsumerIDHeader:
		return "KONG_X_CONSUMER_ID"
	}
	return ""
}

type EndUsers []*DetectedUser

func (ops EndUsers) Find(fn func(op *DetectedUser) bool) (int, *DetectedUser) {
	for i, op := range ops {
		if fn(op) {
			return i, op
		}
	}
	return 0, nil
}

type DetectedUser struct {
	Source    DetectedUserSource `json:"source"`
	ID        string             `json:"id"`
	IPAddress string             `json:"ip_address"`
}

type AuthorizationModel struct {
	Operations Operations `json:"operations"`
}

type Operations []*Operation

func (ops Operations) Find(fn func(op *Operation) bool) (int, *Operation) {
	for i, op := range ops {
		if fn(op) {
			return i, op
		}
	}
	return 0, nil
}

type Audience []*SourceObject

func (aud Audience) Find(fn func(sa *SourceObject) bool) (int, *SourceObject) {
	for i, sa := range aud {
		if fn(sa) {
			return i, sa
		}
	}
	return 0, nil
}
