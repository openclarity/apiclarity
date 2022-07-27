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

package speculatoraccessor

import (
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

type SpeculatorAccessor interface {
	DiffTelemetry(telemetry *_spec.Telemetry, diffSource _spec.DiffSource) (*_spec.APIDiff, error)
	HasApprovedSpec(specKey _speculator.SpecKey) bool
	HasProvidedSpec(specKey _speculator.SpecKey) bool
	GetProvidedSpecVersion(specKey _speculator.SpecKey) _spec.OASVersion
	GetApprovedSpecVersion(specKey _speculator.SpecKey) _spec.OASVersion
}

func NewSpeculatorAccessor(speculator *_speculator.Speculator) SpeculatorAccessor {
	return &Impl{speculator: speculator}
}

type Impl struct {
	speculator *_speculator.Speculator
}

func (s *Impl) DiffTelemetry(telemetry *_spec.Telemetry, diffSource _spec.DiffSource) (*_spec.APIDiff, error) {
	//nolint: wrapcheck
	return s.speculator.DiffTelemetry(telemetry, diffSource)
}

func (s *Impl) HasApprovedSpec(specKey _speculator.SpecKey) bool {
	return s.speculator.HasApprovedSpec(specKey)
}

func (s *Impl) HasProvidedSpec(specKey _speculator.SpecKey) bool {
	return s.speculator.HasProvidedSpec(specKey)
}

func (s *Impl) GetProvidedSpecVersion(specKey _speculator.SpecKey) _spec.OASVersion {
	return s.speculator.GetProvidedSpecVersion(specKey)
}

func (s *Impl) GetApprovedSpecVersion(specKey _speculator.SpecKey) _spec.OASVersion {
	return s.speculator.GetApprovedSpecVersion(specKey)
}
