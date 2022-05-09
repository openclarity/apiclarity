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

package core

const AlertAnnotation = "ALERT"

type AlertSeverity int

const (
	AlertInfo AlertSeverity = iota
	AlertWarn
	AlertCritical
)

var (
	AlertInfoAnn     = Annotation{Name: AlertAnnotation, Annotation: []byte(AlertInfo.String())}
	AlertWarnAnn     = Annotation{Name: AlertAnnotation, Annotation: []byte(AlertWarn.String())}
	AlertCriticalAnn = Annotation{Name: AlertAnnotation, Annotation: []byte(AlertCritical.String())}
)

func (es AlertSeverity) String() string {
	switch es {
	case AlertInfo:
		return "ALERT_INFO"
	case AlertWarn:
		return "ALERT_WARN"
	case AlertCritical:
		return "ALERT_CRITICAL"
	}
	panic("undefined alert severity")
}
