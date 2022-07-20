// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package utils

import (
	"strings"

	"github.com/go-openapi/jsonpointer"
)

func JSONPointer(tokens ...string) string {
	escapedTokens := []string{}
	for _, token := range tokens {
		if token == "" { // stop concatenation of token as soon as a token is empty
			break
		}
		escapedTokens = append(escapedTokens, jsonpointer.Escape(token))
	}

	if len(escapedTokens) == 0 {
		return ""
	}

	return "/" + strings.Join(escapedTokens, "/")
}
