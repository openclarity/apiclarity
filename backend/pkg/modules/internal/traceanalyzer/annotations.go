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

package traceanalyzer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/traceanalyzer/utils"
)


func getAPIDescription(a core.Annotation) utils.Finding {
	switch a.Name {
	case "GUESSABLE_ID":
		var reasons []ParameterFinding
		f := utils.Finding{
			ShortDesc:    "Guessable identifier",
			DetailedDesc: "",
			Severity:     utils.SeverityInfo,
			Alert:        utils.SeverityToAlert(utils.SeverityInfo),
		}
		err := json.Unmarshal(a.Annotation, &reasons)
		if err != nil {
			f.DetailedDesc = fmt.Sprintf("Parameter(s) are guessable")
		} else {
			var descs []string
			for _, r := range reasons {
				descs = append(descs, fmt.Sprintf("%s in '%s %s'", r.Name, r.Method, r.Location))
			}

			f.DetailedDesc = fmt.Sprintf("The following parameter(s) seems to be guessable: %s", strings.Join(descs, ", "))
		}
		return f

	default:
		return utils.Finding{
			ShortDesc:    a.Name,
			DetailedDesc: fmt.Sprintf("[No Description] %s", a.Name),
			Severity:     utils.SeverityInfo,
			Alert:        utils.SeverityToAlert(utils.SeverityInfo),
		}
	}
}
