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
	speculators_repo "github.com/openclarity/apiclarity/backend/pkg/speculators"
	_spec "github.com/openclarity/speculator/pkg/spec"
	_speculator "github.com/openclarity/speculator/pkg/speculator"
)

type SpeculatorsAccessor interface {
	DiffTelemetry(speculatorID uint, telemetry *_spec.Telemetry, diffSource _spec.SpecSource) (*_spec.APIDiff, error)
	HasApprovedSpec(speculatorID uint, specKey _speculator.SpecKey) bool
	HasProvidedSpec(speculatorID uint, specKey _speculator.SpecKey) bool
	GetProvidedSpecVersion(speculatorID uint, specKey _speculator.SpecKey) _spec.OASVersion
	GetApprovedSpecVersion(speculatorID uint, specKey _speculator.SpecKey) _spec.OASVersion
}

func NewSpeculatorAccessor(speculators *speculators_repo.Repository) SpeculatorsAccessor {
	return &Impl{speculators: speculators}
}

type Impl struct {
	speculators *speculators_repo.Repository
}

func (s *Impl) DiffTelemetry(speculatorID uint, telemetry *_spec.Telemetry, diffSource _spec.SpecSource) (*_spec.APIDiff, error) {
	//nolint: wrapcheck
	return s.speculators.Get(speculatorID).DiffTelemetry(telemetry, diffSource)
}

func (s *Impl) HasApprovedSpec(speculatorID uint, specKey _speculator.SpecKey) bool {
	return s.speculators.Get(speculatorID).HasApprovedSpec(specKey)
}

func (s *Impl) HasProvidedSpec(speculatorID uint, specKey _speculator.SpecKey) bool {
	return s.speculators.Get(speculatorID).HasProvidedSpec(specKey)
}

func (s *Impl) GetProvidedSpecVersion(speculatorID uint, specKey _speculator.SpecKey) _spec.OASVersion {
	return s.speculators.Get(speculatorID).GetProvidedSpecVersion(specKey)
}

func (s *Impl) GetApprovedSpecVersion(speculatorID uint, specKey _speculator.SpecKey) _spec.OASVersion {
	return s.speculators.Get(speculatorID).GetApprovedSpecVersion(specKey)
}
