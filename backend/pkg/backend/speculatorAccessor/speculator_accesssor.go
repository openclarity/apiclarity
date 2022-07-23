package speculatorAccessor

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
	return &SpeculatorAccessorImpl{speculator: speculator}
}

type SpeculatorAccessorImpl struct {
	speculator *_speculator.Speculator
}

func (s *SpeculatorAccessorImpl) DiffTelemetry(telemetry *_spec.Telemetry, diffSource _spec.DiffSource) (*_spec.APIDiff, error) {
	return s.speculator.DiffTelemetry(telemetry, diffSource)
}

func (s *SpeculatorAccessorImpl) HasApprovedSpec(specKey _speculator.SpecKey) bool {
	return s.speculator.HasApprovedSpec(specKey)
}

func (s *SpeculatorAccessorImpl) HasProvidedSpec(specKey _speculator.SpecKey) bool {
	return s.speculator.HasProvidedSpec(specKey)
}

func (s *SpeculatorAccessorImpl) GetProvidedSpecVersion(specKey _speculator.SpecKey) _spec.OASVersion {
	return s.speculator.GetProvidedSpecVersion(specKey)
}

func (s *SpeculatorAccessorImpl) GetApprovedSpecVersion(specKey _speculator.SpecKey) _spec.OASVersion {
	return s.speculator.GetApprovedSpecVersion(specKey)
}
