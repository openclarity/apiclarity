// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecParams creates a new DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams object
//
// There are no default values defined in the spec.
func NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecParams() DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams {

	return DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams{}
}

// DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams contains all the bound params for the delete API inventory API ID specs reconstructed spec operation
// typically these are obtained from a http.Request
//
// swagger:parameters DeleteAPIInventoryAPIIDSpecsReconstructedSpec
type DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*
	  Required: true
	  In: path
	*/
	APIID uint32
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecParams() beforehand.
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rAPIID, rhkAPIID, _ := route.Params.GetOK("apiId")
	if err := o.bindAPIID(rAPIID, rhkAPIID, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindAPIID binds and validates parameter APIID from path.
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams) bindAPIID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	value, err := swag.ConvertUint32(raw)
	if err != nil {
		return errors.InvalidType("apiId", "path", "uint32", raw)
	}
	o.APIID = value

	return nil
}
