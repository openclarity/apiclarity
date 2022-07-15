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
	"regexp"
	"time"

	"github.com/go-openapi/spec"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/k8straceannotator"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/bfla/restapi"
)

var spaceRegex = regexp.MustCompile(`\s+`)

type Operation struct {
	Method   string   `json:"method"`
	Path     string   `json:"path"`
	Tags     []string `json:"tags"`
	Audience Audience `json:"audience"`
}

type SourceObject struct {
	K8sObject     *k8straceannotator.K8sObjectRef `json:"k8s_object"`
	External      bool                            `json:"external"`
	EndUsers      EndUsers                        `json:"end_users,omitempty"`
	LastTime      time.Time                       `json:"last_time"`
	StatusCode    int64                           `json:"status_code"`
	WarningStatus restapi.BFLAStatus              `json:"warning_status"`
	Authorized    bool                            `json:"authorized"`
}

func (u *DetectedUser) IsMismatchedScopes(op *spec.Operation) bool {
	if u.Source != DetectedUserSourceJWT || u.JWTClaims == nil {
		return false
	}
	if u.JWTClaims.Scope == nil {
		return false
	}
	for _, secItem := range op.Security {
		for _, scopes := range secItem {
			if !ContainsAll(scopes, spaceRegex.Split(*u.JWTClaims.Scope, -1)) {
				return true
			}
		}
	}
	return false
}

func DetectedUserSourceFromString(s string) DetectedUserSource {
	switch s {
	case "JWT":
		return DetectedUserSourceJWT
	case "BASIC":
		return DetectedUserSourceBasic
	case "KONG_X_CONSUMER_ID":
		return DetectedUserSourceXConsumerIDHeader
	}
	return DetectedUserSourceUnknown
}

type DetectedUserSource int32

const (
	DetectedUserSourceUnknown DetectedUserSource = iota
	DetectedUserSourceJWT
	DetectedUserSourceBasic
	DetectedUserSourceXConsumerIDHeader
)

func (d *DetectedUserSource) UnmarshalJSON(b []byte) error {
	buff := bytes.NewBuffer(b)
	srcName := ""
	fmt.Fscanf(buff, "%q", &srcName)
	*d = DetectedUserSourceFromString(srcName)
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
	case DetectedUserSourceUnknown:
		return ""
	default:
		return ""
	}
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

	// Present if the source is JWT.
	JWTClaims *JWTClaimsWithScopes `json:"jwt_claims"`
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

func ToRestapiSpecType(specType SpecType) restapi.SpecType {
	switch specType {
	case SpecTypeNone:
		return restapi.NONE
	case SpecTypeProvided:
		return restapi.PROVIDED
	case SpecTypeReconstructed:
		return restapi.RECONSTRUCTED
	}
	return restapi.NONE
}
