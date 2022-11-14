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

package utils

import (
	"strings"
	"time"

	oapicommon "github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/utils"
)

type TraceAnalyzerAnnotation interface {
	Name() string
	NewAPIAnnotation(path, method string) TraceAnalyzerAPIAnnotation
	Severity() string
	Redacted() TraceAnalyzerAnnotation
	ToFinding() Finding
}

type TraceAnalyzerAPIAnnotation interface {
	Name() string
	Path() string
	Method() string
	Aggregate(TraceAnalyzerAnnotation) (notify bool)
	Severity() string
	TTL() time.Duration
	Redacted() TraceAnalyzerAPIAnnotation
	ToAPIFinding() oapicommon.APIFinding
}

type BaseTraceAnalyzerAPIAnnotation struct {
	SpecPath   string `json:"path"`
	SpecMethod string `json:"method"`
}

func (a BaseTraceAnalyzerAPIAnnotation) Path() string       { return a.SpecPath }
func (a BaseTraceAnalyzerAPIAnnotation) Method() string     { return a.SpecMethod }
func (a BaseTraceAnalyzerAPIAnnotation) Severity() string   { return SeverityInfo }
func (a BaseTraceAnalyzerAPIAnnotation) TTL() time.Duration { return 24 * time.Hour } //nolint:gomnd
func (a BaseTraceAnalyzerAPIAnnotation) SpecLocation() string {
	if a.SpecPath != "" {
		return utils.JSONPointer("paths", a.SpecPath, strings.ToLower(a.SpecMethod))
	}

	// When no spec is available, the trace analyzer can not specify the
	// location inside the spec. In that case we return an empty location.
	return ""
}

// A finding is an interpreted annotation.
type Finding struct {
	ShortDesc    string
	DetailedDesc string
	Severity     string
	Alert        core.AlertSeverity
}
