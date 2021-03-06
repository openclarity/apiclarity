// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NewGetAPIInventoryAPIIDAPIInfoParams creates a new GetAPIInventoryAPIIDAPIInfoParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAPIInventoryAPIIDAPIInfoParams() *GetAPIInventoryAPIIDAPIInfoParams {
	return &GetAPIInventoryAPIIDAPIInfoParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPIInventoryAPIIDAPIInfoParamsWithTimeout creates a new GetAPIInventoryAPIIDAPIInfoParams object
// with the ability to set a timeout on a request.
func NewGetAPIInventoryAPIIDAPIInfoParamsWithTimeout(timeout time.Duration) *GetAPIInventoryAPIIDAPIInfoParams {
	return &GetAPIInventoryAPIIDAPIInfoParams{
		timeout: timeout,
	}
}

// NewGetAPIInventoryAPIIDAPIInfoParamsWithContext creates a new GetAPIInventoryAPIIDAPIInfoParams object
// with the ability to set a context for a request.
func NewGetAPIInventoryAPIIDAPIInfoParamsWithContext(ctx context.Context) *GetAPIInventoryAPIIDAPIInfoParams {
	return &GetAPIInventoryAPIIDAPIInfoParams{
		Context: ctx,
	}
}

// NewGetAPIInventoryAPIIDAPIInfoParamsWithHTTPClient creates a new GetAPIInventoryAPIIDAPIInfoParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAPIInventoryAPIIDAPIInfoParamsWithHTTPClient(client *http.Client) *GetAPIInventoryAPIIDAPIInfoParams {
	return &GetAPIInventoryAPIIDAPIInfoParams{
		HTTPClient: client,
	}
}

/* GetAPIInventoryAPIIDAPIInfoParams contains all the parameters to send to the API endpoint
   for the get API inventory API ID API info operation.

   Typically these are written to a http.Request.
*/
type GetAPIInventoryAPIIDAPIInfoParams struct {

	// APIID.
	//
	// Format: uint32
	APIID uint32

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get API inventory API ID API info params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIInventoryAPIIDAPIInfoParams) WithDefaults() *GetAPIInventoryAPIIDAPIInfoParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get API inventory API ID API info params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIInventoryAPIIDAPIInfoParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) WithTimeout(timeout time.Duration) *GetAPIInventoryAPIIDAPIInfoParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) WithContext(ctx context.Context) *GetAPIInventoryAPIIDAPIInfoParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) WithHTTPClient(client *http.Client) *GetAPIInventoryAPIIDAPIInfoParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAPIID adds the aPIID to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) WithAPIID(aPIID uint32) *GetAPIInventoryAPIIDAPIInfoParams {
	o.SetAPIID(aPIID)
	return o
}

// SetAPIID adds the apiId to the get API inventory API ID API info params
func (o *GetAPIInventoryAPIIDAPIInfoParams) SetAPIID(aPIID uint32) {
	o.APIID = aPIID
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPIInventoryAPIIDAPIInfoParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param apiId
	if err := r.SetPathParam("apiId", swag.FormatUint32(o.APIID)); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
