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
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
)

const (
	SeverityInfo     = "INFO"
	SeverityLow      = "LOW"
	SeverityMedium   = "MEDIUM"
	SeverityHigh     = "HIGH"
	SeverityCritical = "CRITICAL"
)

func SeverityToAlert(severity string) core.AlertSeverity {
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
