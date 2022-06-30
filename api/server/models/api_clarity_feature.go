// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// APIClarityFeature API clarity feature
//
// swagger:model APIClarityFeature
type APIClarityFeature struct {

	// Short human readable description of the feature
	FeatureDescription string `json:"featureDescription,omitempty"`

	// feature name
	// Required: true
	FeatureName *APIClarityFeatureEnum `json:"featureName"`

	// hosts to trace
	HostsToTrace []string `json:"hostsToTrace"`
}

// Validate validates this API clarity feature
func (m *APIClarityFeature) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateFeatureName(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APIClarityFeature) validateFeatureName(formats strfmt.Registry) error {

	if err := validate.Required("featureName", "body", m.FeatureName); err != nil {
		return err
	}

	if err := validate.Required("featureName", "body", m.FeatureName); err != nil {
		return err
	}

	if m.FeatureName != nil {
		if err := m.FeatureName.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("featureName")
			}
			return err
		}
	}

	return nil
}

// ContextValidate validate this API clarity feature based on the context it is used
func (m *APIClarityFeature) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateFeatureName(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *APIClarityFeature) contextValidateFeatureName(ctx context.Context, formats strfmt.Registry) error {

	if m.FeatureName != nil {
		if err := m.FeatureName.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("featureName")
			}
			return err
		}
	}

	return nil
}

// MarshalBinary interface implementation
func (m *APIClarityFeature) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *APIClarityFeature) UnmarshalBinary(b []byte) error {
	var res APIClarityFeature
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
