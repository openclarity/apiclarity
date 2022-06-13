// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// ApprovedReview approved review
//
// swagger:model ApprovedReview
type ApprovedReview struct {

	// OpenAPI version to use when saving the approved spec
	// Enum: [OASv2.0 OASv3.0]
	OasVersion string `json:"oasVersion,omitempty"`

	// review path items
	ReviewPathItems []*ReviewPathItem `json:"reviewPathItems"`
}

// Validate validates this approved review
func (m *ApprovedReview) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateOasVersion(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateReviewPathItems(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

var approvedReviewTypeOasVersionPropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["OASv2.0","OASv3.0"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		approvedReviewTypeOasVersionPropEnum = append(approvedReviewTypeOasVersionPropEnum, v)
	}
}

const (

	// ApprovedReviewOasVersionOASv2Dot0 captures enum value "OASv2.0"
	ApprovedReviewOasVersionOASv2Dot0 string = "OASv2.0"

	// ApprovedReviewOasVersionOASv3Dot0 captures enum value "OASv3.0"
	ApprovedReviewOasVersionOASv3Dot0 string = "OASv3.0"
)

// prop value enum
func (m *ApprovedReview) validateOasVersionEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, approvedReviewTypeOasVersionPropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *ApprovedReview) validateOasVersion(formats strfmt.Registry) error {
	if swag.IsZero(m.OasVersion) { // not required
		return nil
	}

	// value enum
	if err := m.validateOasVersionEnum("oasVersion", "body", m.OasVersion); err != nil {
		return err
	}

	return nil
}

func (m *ApprovedReview) validateReviewPathItems(formats strfmt.Registry) error {
	if swag.IsZero(m.ReviewPathItems) { // not required
		return nil
	}

	for i := 0; i < len(m.ReviewPathItems); i++ {
		if swag.IsZero(m.ReviewPathItems[i]) { // not required
			continue
		}

		if m.ReviewPathItems[i] != nil {
			if err := m.ReviewPathItems[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("reviewPathItems" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this approved review based on the context it is used
func (m *ApprovedReview) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateReviewPathItems(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *ApprovedReview) contextValidateReviewPathItems(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.ReviewPathItems); i++ {

		if m.ReviewPathItems[i] != nil {
			if err := m.ReviewPathItems[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("reviewPathItems" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *ApprovedReview) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ApprovedReview) UnmarshalBinary(b []byte) error {
	var res ApprovedReview
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
