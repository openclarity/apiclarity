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

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

const (
	SeverityInfo     = "INFO"
	SeverityWarn     = "WARN"
	SeverityLow      = "LOW"
	SeverityMedium   = "MEDIUM"
	SeverityHigh     = "HIGH"
	SeverityCritical = "CRITICAL"
)

func getEventDescription(a core.Annotation) Finding {
	switch a.Name {
	case "BASIC_AUTH_SHORT_PASSWORD":
		return Finding{
			ShortDesc:    "Too short Basic Auth password",
			DetailedDesc: fmt.Sprintf("The the length of Basic Auth password is too short (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "BASIC_AUTH_KNOWN_PASSWORD":
		return Finding{
			ShortDesc:    "Weak Basic Auth password (found in dictionary)",
			DetailedDesc: fmt.Sprintf("The Basic Auth password is too weak because it's too common (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "BASIC_AUTH_SAME_PASSWORD":
		p := bytes.Split(a.Annotation, []byte{','})
		f := Finding{
			ShortDesc: "Same Basic Auth credentials used for another service",
			Severity:  SeverityMedium,
			Alert:     &core.AlertInfoAnn,
		}
		//nolint:gomnd
		if len(p) < 3 {
			f.DetailedDesc = "The exact same Basic Auth credentials of this event are used for another service"
		} else {
			f.DetailedDesc = fmt.Sprintf("The exact same Basic Auth credentials (%s) of this event are used for multiple services (%s)", p[0], bytes.Join(p[1:], []byte{','}))
		}
		return f

	case "REGEXP_MATCHING_REQUEST_BODY":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains a sensitive information in the request body (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "REGEXP_MATCHING_RESPONSE_BODY":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains a sensitive information in the response body (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "REGEXP_MATCHING_REQUEST_HEADERS":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains a sensitive information in the request headers (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "REGEXP_MATCHING_RESPONSE_HEADERS":
		return Finding{
			ShortDesc:    "Matching regular expression",
			DetailedDesc: fmt.Sprintf("This event contains a sensitive information in the response headers (%s)", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}

	case "JWT_WEAK_SYMETRIC_SECRET":
		return Finding{
			ShortDesc:    "JWT signed with a weak key",
			DetailedDesc: fmt.Sprintf("The weak signing key is '%s'", a.Annotation),
			Severity:     SeverityHigh,
			Alert:        &core.AlertWarnAnn,
		}
	case "JWT_NO_ALG_FIELD":
		return Finding{
			ShortDesc:    "JWT has no algorithm specificed",
			DetailedDesc: "The JOSE header of the JWT header does not contain an 'alg' field",
			Severity:     SeverityHigh,
			Alert:        &core.AlertCriticalAnn,
		}
	case "JWT_NOT_RECOMMENDED_ALG":
		return Finding{
			ShortDesc:    "Not a recommanded JWT signing algorithm",
			DetailedDesc: fmt.Sprintf("'%s' is not a recommended signing algorithm", a.Annotation),
			Severity:     SeverityHigh,
			Alert:        &core.AlertInfoAnn,
		}
	case "JWT_EXP_TOO_FAR":
		return Finding{
			ShortDesc:    "JWT expire too far in the future",
			DetailedDesc: string(a.Annotation),
			Severity:     SeverityLow,
			Alert:        nil,
		}
	case "JWT_NO_EXPIRE_CLAIM":
		return Finding{
			ShortDesc:    "JWT does not have any expire claims",
			DetailedDesc: string(a.Annotation),
			Severity:     SeverityLow,
			Alert:        nil,
		}
	case "JWT_SENSITIVE_CONTENT_IN_CLAIMS":
		return Finding{
			ShortDesc:    "JWT claims may contains sensitive content",
			DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "JWT_SENSITIVE_CONTENT_IN_HEADERS":
		return Finding{
			ShortDesc:    "JWT headers may contain sensitive content",
			DetailedDesc: fmt.Sprintf("JWT are signed, not encrypted, hence sensitive information can be seen in clear by a potential attacker. Here, '%s' seems sensitive", a.Annotation),
			Severity:     SeverityMedium,
			Alert:        nil,
		}
	case "NLID":
		var reason ParameterFinding
		f := Finding{
			ShortDesc:    "NLID (Non learnt Identifier)",
			DetailedDesc: "",
			Severity:     SeverityInfo,
			Alert:        nil,
		}
		err := json.Unmarshal(a.Annotation, &reason)
		if err != nil {
			f.DetailedDesc = "A parameter was used that was not previously retrieved. Potential BOLA."
		} else {
			f.DetailedDesc = fmt.Sprintf("A Parameter in '%s %s' was used but not previously retrieved. Potential BOLA.", reason.Method, reason.Location)
		}
		return f
	default:
		return Finding{
			ShortDesc:    a.Name,
			DetailedDesc: fmt.Sprintf("[No Description] %s", a.Name),
			Severity:     SeverityInfo,
			Alert:        nil,
		}
	}
}

func getAPIDescription(a core.Annotation) Finding {
	switch a.Name {
	case "GUESSABLE_ID":
		var reason ParameterFinding
		f := Finding{
			ShortDesc:    "Guessable identifier",
			DetailedDesc: "",
			Severity:     SeverityInfo,
			Alert:        &core.AlertInfoAnn,
		}
		err := json.Unmarshal(a.Annotation, &reason)
		if err != nil {
			f.DetailedDesc = "A parameter is guessable"
		} else {
			f.DetailedDesc = fmt.Sprintf("Parameter '%s' in '%s %s' seems to be guessable", reason.Name, reason.Method, reason.Location)
		}
		return f

	default:
		return Finding{
			ShortDesc:    a.Name,
			DetailedDesc: fmt.Sprintf("[No Description] %s", a.Name),
			Severity:     SeverityInfo,
			Alert:        nil,
		}
	}
}
