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
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

const (
	SeverityInfo     = "INFO"
	SeverityLow      = "LOW"
	SeverityMedium   = "MEDIUM"
	SeverityHigh     = "HIGH"
	SeverityCritical = "CRITICAL"
)

func severityToAlert(severity string) core.AlertSeverity {
	switch severity {
	case SeverityInfo, SeverityLow:
		return core.AlertInfo
	case SeverityMedium, SeverityHigh:
		return core.AlertWarn
	case SeverityCritical:
		return core.AlertCritical
	}

	return core.AlertInfo
}

func getEventDescription(a core.Annotation) Finding {
	switch a.Name {
	case "BASIC_AUTH_SHORT_PASSWORD":
		return Finding{
			ShortDesc:    "Too short Basic Auth password",
			DetailedDesc: fmt.Sprintf("The the length of Basic Auth password is too short (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "BASIC_AUTH_KNOWN_PASSWORD":
		return Finding{
			ShortDesc:    "Weak Basic Auth password (found in dictionary)",
			DetailedDesc: fmt.Sprintf("The Basic Auth password is too weak because it's too common (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "BASIC_AUTH_SAME_PASSWORD":
		p := bytes.Split(a.Annotation, []byte{','})
		f := Finding{
			ShortDesc: "Same Basic Auth credentials used for another service",
			Severity:  SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
		if len(p) < 3 {
			f.DetailedDesc = "The exact same Basic Auth credentials of this event are used for another service"
		} else {
			f.DetailedDesc = fmt.Sprintf("The exact same Basic Auth credentials (%s) of this event are used for multiple services (%s)", p[0], bytes.Join(p[1:], []byte{','}))
		}
		return f

	case "REGEXP_MATCHING_REQUEST_BODY":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains sensitive information in the request body (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "REGEXP_MATCHING_RESPONSE_BODY":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains sensitive information in the response body (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "REGEXP_MATCHING_REQUEST_HEADERS":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains sensitive information in the request headers (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "REGEXP_MATCHING_RESPONSE_HEADERS":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains sensitive information in the response headers (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}

	case "JWT_WEAK_SYMETRIC_SECRET":
		return Finding{
			ShortDesc:    "JWT signed with a weak key",
			DetailedDesc: fmt.Sprintf("The weak signing key is '%s'", a.Annotation),
			Severity:     SeverityHigh,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "JWT_NO_ALG_FIELD":
		return Finding{
			ShortDesc:    "JWT has no algorithm specificed",
			DetailedDesc: fmt.Sprintf("The JOSE header of the JWT header does not contain an 'alg' field"),
			Severity:     SeverityHigh,
			Alert:        severityToAlert(SeverityHigh),
		}
	case "JWT_NOT_RECOMMENDED_ALG":
		return Finding{
			ShortDesc:    "Not a recommanded JWT signing algorithm",
			DetailedDesc: fmt.Sprintf("'%s' is not a recommended signing algorithm", a.Annotation),
			Severity:     SeverityHigh,
			Alert:        severityToAlert(SeverityHigh),
		}
	case "JWT_EXP_TOO_FAR":
		return Finding{
			ShortDesc:    "JWT expire too far in the future",
			DetailedDesc: string(a.Annotation),
			Severity:     SeverityLow,
			Alert:        severityToAlert(SeverityLow),
		}
	case "JWT_NO_EXPIRE_CLAIM":
		return Finding{
			ShortDesc:    "JWT does not have any expire claims",
			DetailedDesc: string(a.Annotation),
			Severity:     SeverityLow,
			Alert:        severityToAlert(SeverityLow),
		}
	case "JWT_SENSITIVE_CONTENT_IN_CLAIMS":
		return Finding{
			ShortDesc:    "JWT claims may contains sensitive content",
			DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "JWT_SENSITIVE_CONTENT_IN_HEADERS":
		return Finding{
			ShortDesc:    "JWT headers may contain sensitive content",
			DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        severityToAlert(SeverityMedium),
		}
	case "NLID":
		var reasons []ParameterFinding
		f := Finding{
			ShortDesc:    "NLID (Non learnt Identifier)",
			DetailedDesc: "",
			Severity:     SeverityInfo,
			Alert:        severityToAlert(SeverityInfo),
		}
		err := json.Unmarshal(a.Annotation, &reasons)
		if err != nil {
			f.DetailedDesc = "Parameter(s) were used but not previously retrieved. Potential BOLA."
		} else {
			var descs []string
			for _, r := range reasons {
				descs = append(descs, fmt.Sprintf("%s %s: %s", r.Method, r.Location, r.Value))
			}

			f.DetailedDesc = fmt.Sprintf("Parameter(s) were used but not previously retrieved. Potential BOLA. (%s)", strings.Join(descs, ", "))
		}
		return f
	default:
		return Finding{
			ShortDesc:    a.Name,
			DetailedDesc: fmt.Sprintf("[No Description] %s", a.Name),
			Severity:     SeverityInfo,
			Alert:        severityToAlert(SeverityInfo),
		}
	}
}

func getAPIDescription(a core.Annotation) Finding {
	switch a.Name {
	case "GUESSABLE_ID":
		var reasons []ParameterFinding
		f := Finding{
			ShortDesc:    "Guessable identifier",
			DetailedDesc: "",
			Severity:     SeverityInfo,
			Alert:        severityToAlert(SeverityInfo),
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
		return Finding{
			ShortDesc:    a.Name,
			DetailedDesc: fmt.Sprintf("[No Description] %s", a.Name),
			Severity:     SeverityInfo,
			Alert:        severityToAlert(SeverityInfo),
		}
	}
}
