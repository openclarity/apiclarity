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

// NewGetAPIInventoryAPIIDSuggestedReviewParams creates a new GetAPIInventoryAPIIDSuggestedReviewParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAPIInventoryAPIIDSuggestedReviewParams() *GetAPIInventoryAPIIDSuggestedReviewParams {
	return &GetAPIInventoryAPIIDSuggestedReviewParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPIInventoryAPIIDSuggestedReviewParamsWithTimeout creates a new GetAPIInventoryAPIIDSuggestedReviewParams object
// with the ability to set a timeout on a request.
func NewGetAPIInventoryAPIIDSuggestedReviewParamsWithTimeout(timeout time.Duration) *GetAPIInventoryAPIIDSuggestedReviewParams {
	return &GetAPIInventoryAPIIDSuggestedReviewParams{
		timeout: timeout,
	}
}

// NewGetAPIInventoryAPIIDSuggestedReviewParamsWithContext creates a new GetAPIInventoryAPIIDSuggestedReviewParams object
// with the ability to set a context for a request.
func NewGetAPIInventoryAPIIDSuggestedReviewParamsWithContext(ctx context.Context) *GetAPIInventoryAPIIDSuggestedReviewParams {
	return &GetAPIInventoryAPIIDSuggestedReviewParams{
		Context: ctx,
	}
}

// NewGetAPIInventoryAPIIDSuggestedReviewParamsWithHTTPClient creates a new GetAPIInventoryAPIIDSuggestedReviewParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAPIInventoryAPIIDSuggestedReviewParamsWithHTTPClient(client *http.Client) *GetAPIInventoryAPIIDSuggestedReviewParams {
	return &GetAPIInventoryAPIIDSuggestedReviewParams{
		HTTPClient: client,
	}
}

/* GetAPIInventoryAPIIDSuggestedReviewParams contains all the parameters to send to the API endpoint
   for the get API inventory API ID suggested review operation.

   Typically these are written to a http.Request.
*/
type GetAPIInventoryAPIIDSuggestedReviewParams struct {

	// APIID.
	//
	// Format: uint32
	APIID uint32

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get API inventory API ID suggested review params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) WithDefaults() *GetAPIInventoryAPIIDSuggestedReviewParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get API inventory API ID suggested review params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) WithTimeout(timeout time.Duration) *GetAPIInventoryAPIIDSuggestedReviewParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) WithContext(ctx context.Context) *GetAPIInventoryAPIIDSuggestedReviewParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) WithHTTPClient(client *http.Client) *GetAPIInventoryAPIIDSuggestedReviewParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAPIID adds the aPIID to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) WithAPIID(aPIID uint32) *GetAPIInventoryAPIIDSuggestedReviewParams {
	o.SetAPIID(aPIID)
	return o
}

// SetAPIID adds the apiId to the get API inventory API ID suggested review params
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) SetAPIID(aPIID uint32) {
	o.APIID = aPIID
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPIInventoryAPIIDSuggestedReviewParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
