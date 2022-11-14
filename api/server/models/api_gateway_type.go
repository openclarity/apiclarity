// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
)

// APIGatewayType API gateway type
//
// swagger:model APIGatewayType
type APIGatewayType string

func NewAPIGatewayType(value APIGatewayType) *APIGatewayType {
	v := value
	return &v
}

const (

	// APIGatewayTypeTYK captures enum value "TYK"
	APIGatewayTypeTYK APIGatewayType = "TYK"

	// APIGatewayTypeKONG captures enum value "KONG"
	APIGatewayTypeKONG APIGatewayType = "KONG"

	// APIGatewayTypeAPIGEEX captures enum value "APIGEEX"
	APIGatewayTypeAPIGEEX APIGatewayType = "APIGEEX"
)

// for schema
var apiGatewayTypeEnum []interface{}

func init() {
	var res []APIGatewayType
	if err := json.Unmarshal([]byte(`["TYK","KONG","APIGEEX"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		apiGatewayTypeEnum = append(apiGatewayTypeEnum, v)
	}
}

func (m APIGatewayType) validateAPIGatewayTypeEnum(path, location string, value APIGatewayType) error {
	if err := validate.EnumCase(path, location, value, apiGatewayTypeEnum, true); err != nil {
		return err
	}
	return nil
}

// Validate validates this API gateway type
func (m APIGatewayType) Validate(formats strfmt.Registry) error {
	var res []error

	// value enum
	if err := m.validateAPIGatewayTypeEnum("", "body", m); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validates this API gateway type based on context it is used
func (m APIGatewayType) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}
