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
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

func GetUserID(headers http.Header) (*DetectedUser, error) {
	if xcustomerID := headers.Get("x-customer-id"); xcustomerID != "" {
		return &DetectedUser{Source: DetectedUserSourceXConsumerIDHeader, ID: xcustomerID}, nil
	}
	authz := headers.Get("authorization")
	if authz == "" {
		return nil, nil
	}
	if strings.HasPrefix(authz, "Basic ") {
		basic := strings.TrimPrefix(authz, "Basic ")
		usernameAndPassword, err := base64.StdEncoding.DecodeString(basic)
		if err != nil {
			return nil, fmt.Errorf("cannot decode basic authz header: %w", err)
		}
		usernameAndPasswordParts := strings.Split(string(usernameAndPassword), ":")

		// nolint:gomnd
		if len(usernameAndPasswordParts) < 2 {
			return nil, errors.New("broken basic auth header")
		}
		return &DetectedUser{Source: DetectedUserSourceBasic, ID: usernameAndPasswordParts[0]}, nil
	}
	if strings.HasPrefix(authz, "Bearer ") {
		bearer := strings.TrimPrefix(authz, "Bearer ")
		claims := &JWTClaimsWithScopes{}
		if _, _, err := jwt.NewParser(jwt.WithoutClaimsValidation()).ParseUnverified(bearer, claims); err != nil {
			return nil, fmt.Errorf("unsuported bearer token: %w", err)
		}
		return &DetectedUser{Source: DetectedUserSourceJWT, ID: claims.Subject, JWTClaims: claims}, nil
	}
	return nil, ErrUnsupportedAuthScheme
}

type JWTClaimsWithScopes struct {
	*jwt.RegisteredClaims

	Scope *string `json:"scope"`
}
