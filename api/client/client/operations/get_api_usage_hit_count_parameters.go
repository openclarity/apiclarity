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

// NewGetAPIUsageHitCountParams creates a new GetAPIUsageHitCountParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetAPIUsageHitCountParams() *GetAPIUsageHitCountParams {
	return &GetAPIUsageHitCountParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetAPIUsageHitCountParamsWithTimeout creates a new GetAPIUsageHitCountParams object
// with the ability to set a timeout on a request.
func NewGetAPIUsageHitCountParamsWithTimeout(timeout time.Duration) *GetAPIUsageHitCountParams {
	return &GetAPIUsageHitCountParams{
		timeout: timeout,
	}
}

// NewGetAPIUsageHitCountParamsWithContext creates a new GetAPIUsageHitCountParams object
// with the ability to set a context for a request.
func NewGetAPIUsageHitCountParamsWithContext(ctx context.Context) *GetAPIUsageHitCountParams {
	return &GetAPIUsageHitCountParams{
		Context: ctx,
	}
}

// NewGetAPIUsageHitCountParamsWithHTTPClient creates a new GetAPIUsageHitCountParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetAPIUsageHitCountParamsWithHTTPClient(client *http.Client) *GetAPIUsageHitCountParams {
	return &GetAPIUsageHitCountParams{
		HTTPClient: client,
	}
}

/* GetAPIUsageHitCountParams contains all the parameters to send to the API endpoint
   for the get API usage hit count operation.

   Typically these are written to a http.Request.
*/
type GetAPIUsageHitCountParams struct {

	// DestinationIPIsNot.
	DestinationIPIsNot []string

	// DestinationIPIs.
	DestinationIPIs []string

	// DestinationPortIsNot.
	DestinationPortIsNot []string

	// DestinationPortIs.
	DestinationPortIs []string

	/* EndTime.

	   End time of the query

	   Format: date-time
	*/
	EndTime strfmt.DateTime

	// HasSpecDiffIs.
	HasSpecDiffIs *bool

	// MethodIs.
	MethodIs []string

	// PathContains.
	PathContains []string

	// PathEnd.
	PathEnd *string

	// PathIsNot.
	PathIsNot []string

	// PathIs.
	PathIs []string

	// PathStart.
	PathStart *string

	// ProvidedPathIDIs.
	ProvidedPathIDIs []string

	// ReconstructedPathIDIs.
	ReconstructedPathIDIs []string

	// ShowNonAPI.
	ShowNonAPI bool

	// SourceIPIsNot.
	SourceIPIsNot []string

	// SourceIPIs.
	SourceIPIs []string

	// SpecDiffTypeIs.
	SpecDiffTypeIs []string

	// SpecContains.
	SpecContains []string

	// SpecEnd.
	SpecEnd *string

	// SpecIsNot.
	SpecIsNot []string

	// SpecIs.
	SpecIs []string

	// SpecStart.
	SpecStart *string

	/* StartTime.

	   Start time of the query

	   Format: date-time
	*/
	StartTime strfmt.DateTime

	/* StatusCodeGte.

	   greater than or equal
	*/
	StatusCodeGte *string

	// StatusCodeIsNot.
	StatusCodeIsNot []string

	// StatusCodeIs.
	StatusCodeIs []string

	/* StatusCodeLte.

	   less than or equal
	*/
	StatusCodeLte *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get API usage hit count params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIUsageHitCountParams) WithDefaults() *GetAPIUsageHitCountParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get API usage hit count params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetAPIUsageHitCountParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithTimeout(timeout time.Duration) *GetAPIUsageHitCountParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithContext(ctx context.Context) *GetAPIUsageHitCountParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithHTTPClient(client *http.Client) *GetAPIUsageHitCountParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithDestinationIPIsNot adds the destinationIPIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithDestinationIPIsNot(destinationIPIsNot []string) *GetAPIUsageHitCountParams {
	o.SetDestinationIPIsNot(destinationIPIsNot)
	return o
}

// SetDestinationIPIsNot adds the destinationIpIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetDestinationIPIsNot(destinationIPIsNot []string) {
	o.DestinationIPIsNot = destinationIPIsNot
}

// WithDestinationIPIs adds the destinationIPIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithDestinationIPIs(destinationIPIs []string) *GetAPIUsageHitCountParams {
	o.SetDestinationIPIs(destinationIPIs)
	return o
}

// SetDestinationIPIs adds the destinationIpIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetDestinationIPIs(destinationIPIs []string) {
	o.DestinationIPIs = destinationIPIs
}

// WithDestinationPortIsNot adds the destinationPortIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithDestinationPortIsNot(destinationPortIsNot []string) *GetAPIUsageHitCountParams {
	o.SetDestinationPortIsNot(destinationPortIsNot)
	return o
}

// SetDestinationPortIsNot adds the destinationPortIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetDestinationPortIsNot(destinationPortIsNot []string) {
	o.DestinationPortIsNot = destinationPortIsNot
}

// WithDestinationPortIs adds the destinationPortIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithDestinationPortIs(destinationPortIs []string) *GetAPIUsageHitCountParams {
	o.SetDestinationPortIs(destinationPortIs)
	return o
}

// SetDestinationPortIs adds the destinationPortIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetDestinationPortIs(destinationPortIs []string) {
	o.DestinationPortIs = destinationPortIs
}

// WithEndTime adds the endTime to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithEndTime(endTime strfmt.DateTime) *GetAPIUsageHitCountParams {
	o.SetEndTime(endTime)
	return o
}

// SetEndTime adds the endTime to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetEndTime(endTime strfmt.DateTime) {
	o.EndTime = endTime
}

// WithHasSpecDiffIs adds the hasSpecDiffIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithHasSpecDiffIs(hasSpecDiffIs *bool) *GetAPIUsageHitCountParams {
	o.SetHasSpecDiffIs(hasSpecDiffIs)
	return o
}

// SetHasSpecDiffIs adds the hasSpecDiffIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetHasSpecDiffIs(hasSpecDiffIs *bool) {
	o.HasSpecDiffIs = hasSpecDiffIs
}

// WithMethodIs adds the methodIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithMethodIs(methodIs []string) *GetAPIUsageHitCountParams {
	o.SetMethodIs(methodIs)
	return o
}

// SetMethodIs adds the methodIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetMethodIs(methodIs []string) {
	o.MethodIs = methodIs
}

// WithPathContains adds the pathContains to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithPathContains(pathContains []string) *GetAPIUsageHitCountParams {
	o.SetPathContains(pathContains)
	return o
}

// SetPathContains adds the pathContains to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetPathContains(pathContains []string) {
	o.PathContains = pathContains
}

// WithPathEnd adds the pathEnd to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithPathEnd(pathEnd *string) *GetAPIUsageHitCountParams {
	o.SetPathEnd(pathEnd)
	return o
}

// SetPathEnd adds the pathEnd to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetPathEnd(pathEnd *string) {
	o.PathEnd = pathEnd
}

// WithPathIsNot adds the pathIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithPathIsNot(pathIsNot []string) *GetAPIUsageHitCountParams {
	o.SetPathIsNot(pathIsNot)
	return o
}

// SetPathIsNot adds the pathIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetPathIsNot(pathIsNot []string) {
	o.PathIsNot = pathIsNot
}

// WithPathIs adds the pathIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithPathIs(pathIs []string) *GetAPIUsageHitCountParams {
	o.SetPathIs(pathIs)
	return o
}

// SetPathIs adds the pathIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetPathIs(pathIs []string) {
	o.PathIs = pathIs
}

// WithPathStart adds the pathStart to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithPathStart(pathStart *string) *GetAPIUsageHitCountParams {
	o.SetPathStart(pathStart)
	return o
}

// SetPathStart adds the pathStart to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetPathStart(pathStart *string) {
	o.PathStart = pathStart
}

// WithProvidedPathIDIs adds the providedPathIDIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithProvidedPathIDIs(providedPathIDIs []string) *GetAPIUsageHitCountParams {
	o.SetProvidedPathIDIs(providedPathIDIs)
	return o
}

// SetProvidedPathIDIs adds the providedPathIdIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetProvidedPathIDIs(providedPathIDIs []string) {
	o.ProvidedPathIDIs = providedPathIDIs
}

// WithReconstructedPathIDIs adds the reconstructedPathIDIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithReconstructedPathIDIs(reconstructedPathIDIs []string) *GetAPIUsageHitCountParams {
	o.SetReconstructedPathIDIs(reconstructedPathIDIs)
	return o
}

// SetReconstructedPathIDIs adds the reconstructedPathIdIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetReconstructedPathIDIs(reconstructedPathIDIs []string) {
	o.ReconstructedPathIDIs = reconstructedPathIDIs
}

// WithShowNonAPI adds the showNonAPI to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithShowNonAPI(showNonAPI bool) *GetAPIUsageHitCountParams {
	o.SetShowNonAPI(showNonAPI)
	return o
}

// SetShowNonAPI adds the showNonApi to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetShowNonAPI(showNonAPI bool) {
	o.ShowNonAPI = showNonAPI
}

// WithSourceIPIsNot adds the sourceIPIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSourceIPIsNot(sourceIPIsNot []string) *GetAPIUsageHitCountParams {
	o.SetSourceIPIsNot(sourceIPIsNot)
	return o
}

// SetSourceIPIsNot adds the sourceIpIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSourceIPIsNot(sourceIPIsNot []string) {
	o.SourceIPIsNot = sourceIPIsNot
}

// WithSourceIPIs adds the sourceIPIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSourceIPIs(sourceIPIs []string) *GetAPIUsageHitCountParams {
	o.SetSourceIPIs(sourceIPIs)
	return o
}

// SetSourceIPIs adds the sourceIpIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSourceIPIs(sourceIPIs []string) {
	o.SourceIPIs = sourceIPIs
}

// WithSpecDiffTypeIs adds the specDiffTypeIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSpecDiffTypeIs(specDiffTypeIs []string) *GetAPIUsageHitCountParams {
	o.SetSpecDiffTypeIs(specDiffTypeIs)
	return o
}

// SetSpecDiffTypeIs adds the specDiffTypeIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSpecDiffTypeIs(specDiffTypeIs []string) {
	o.SpecDiffTypeIs = specDiffTypeIs
}

// WithSpecContains adds the specContains to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSpecContains(specContains []string) *GetAPIUsageHitCountParams {
	o.SetSpecContains(specContains)
	return o
}

// SetSpecContains adds the specContains to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSpecContains(specContains []string) {
	o.SpecContains = specContains
}

// WithSpecEnd adds the specEnd to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSpecEnd(specEnd *string) *GetAPIUsageHitCountParams {
	o.SetSpecEnd(specEnd)
	return o
}

// SetSpecEnd adds the specEnd to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSpecEnd(specEnd *string) {
	o.SpecEnd = specEnd
}

// WithSpecIsNot adds the specIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSpecIsNot(specIsNot []string) *GetAPIUsageHitCountParams {
	o.SetSpecIsNot(specIsNot)
	return o
}

// SetSpecIsNot adds the specIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSpecIsNot(specIsNot []string) {
	o.SpecIsNot = specIsNot
}

// WithSpecIs adds the specIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSpecIs(specIs []string) *GetAPIUsageHitCountParams {
	o.SetSpecIs(specIs)
	return o
}

// SetSpecIs adds the specIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSpecIs(specIs []string) {
	o.SpecIs = specIs
}

// WithSpecStart adds the specStart to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithSpecStart(specStart *string) *GetAPIUsageHitCountParams {
	o.SetSpecStart(specStart)
	return o
}

// SetSpecStart adds the specStart to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetSpecStart(specStart *string) {
	o.SpecStart = specStart
}

// WithStartTime adds the startTime to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithStartTime(startTime strfmt.DateTime) *GetAPIUsageHitCountParams {
	o.SetStartTime(startTime)
	return o
}

// SetStartTime adds the startTime to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetStartTime(startTime strfmt.DateTime) {
	o.StartTime = startTime
}

// WithStatusCodeGte adds the statusCodeGte to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithStatusCodeGte(statusCodeGte *string) *GetAPIUsageHitCountParams {
	o.SetStatusCodeGte(statusCodeGte)
	return o
}

// SetStatusCodeGte adds the statusCodeGte to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetStatusCodeGte(statusCodeGte *string) {
	o.StatusCodeGte = statusCodeGte
}

// WithStatusCodeIsNot adds the statusCodeIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithStatusCodeIsNot(statusCodeIsNot []string) *GetAPIUsageHitCountParams {
	o.SetStatusCodeIsNot(statusCodeIsNot)
	return o
}

// SetStatusCodeIsNot adds the statusCodeIsNot to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetStatusCodeIsNot(statusCodeIsNot []string) {
	o.StatusCodeIsNot = statusCodeIsNot
}

// WithStatusCodeIs adds the statusCodeIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithStatusCodeIs(statusCodeIs []string) *GetAPIUsageHitCountParams {
	o.SetStatusCodeIs(statusCodeIs)
	return o
}

// SetStatusCodeIs adds the statusCodeIs to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetStatusCodeIs(statusCodeIs []string) {
	o.StatusCodeIs = statusCodeIs
}

// WithStatusCodeLte adds the statusCodeLte to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) WithStatusCodeLte(statusCodeLte *string) *GetAPIUsageHitCountParams {
	o.SetStatusCodeLte(statusCodeLte)
	return o
}

// SetStatusCodeLte adds the statusCodeLte to the get API usage hit count params
func (o *GetAPIUsageHitCountParams) SetStatusCodeLte(statusCodeLte *string) {
	o.StatusCodeLte = statusCodeLte
}

// WriteToRequest writes these params to a swagger request
func (o *GetAPIUsageHitCountParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.DestinationIPIsNot != nil {

		// binding items for destinationIP[isNot]
		joinedDestinationIPIsNot := o.bindParamDestinationIPIsNot(reg)

		// query array param destinationIP[isNot]
		if err := r.SetQueryParam("destinationIP[isNot]", joinedDestinationIPIsNot...); err != nil {
			return err
		}
	}

	if o.DestinationIPIs != nil {

		// binding items for destinationIP[is]
		joinedDestinationIPIs := o.bindParamDestinationIPIs(reg)

		// query array param destinationIP[is]
		if err := r.SetQueryParam("destinationIP[is]", joinedDestinationIPIs...); err != nil {
			return err
		}
	}

	if o.DestinationPortIsNot != nil {

		// binding items for destinationPort[isNot]
		joinedDestinationPortIsNot := o.bindParamDestinationPortIsNot(reg)

		// query array param destinationPort[isNot]
		if err := r.SetQueryParam("destinationPort[isNot]", joinedDestinationPortIsNot...); err != nil {
			return err
		}
	}

	if o.DestinationPortIs != nil {

		// binding items for destinationPort[is]
		joinedDestinationPortIs := o.bindParamDestinationPortIs(reg)

		// query array param destinationPort[is]
		if err := r.SetQueryParam("destinationPort[is]", joinedDestinationPortIs...); err != nil {
			return err
		}
	}

	// query param endTime
	qrEndTime := o.EndTime
	qEndTime := qrEndTime.String()
	if qEndTime != "" {

		if err := r.SetQueryParam("endTime", qEndTime); err != nil {
			return err
		}
	}

	if o.HasSpecDiffIs != nil {

		// query param hasSpecDiff[is]
		var qrHasSpecDiffIs bool

		if o.HasSpecDiffIs != nil {
			qrHasSpecDiffIs = *o.HasSpecDiffIs
		}
		qHasSpecDiffIs := swag.FormatBool(qrHasSpecDiffIs)
		if qHasSpecDiffIs != "" {

			if err := r.SetQueryParam("hasSpecDiff[is]", qHasSpecDiffIs); err != nil {
				return err
			}
		}
	}

	if o.MethodIs != nil {

		// binding items for method[is]
		joinedMethodIs := o.bindParamMethodIs(reg)

		// query array param method[is]
		if err := r.SetQueryParam("method[is]", joinedMethodIs...); err != nil {
			return err
		}
	}

	if o.PathContains != nil {

		// binding items for path[contains]
		joinedPathContains := o.bindParamPathContains(reg)

		// query array param path[contains]
		if err := r.SetQueryParam("path[contains]", joinedPathContains...); err != nil {
			return err
		}
	}

	if o.PathEnd != nil {

		// query param path[end]
		var qrPathEnd string

		if o.PathEnd != nil {
			qrPathEnd = *o.PathEnd
		}
		qPathEnd := qrPathEnd
		if qPathEnd != "" {

			if err := r.SetQueryParam("path[end]", qPathEnd); err != nil {
				return err
			}
		}
	}

	if o.PathIsNot != nil {

		// binding items for path[isNot]
		joinedPathIsNot := o.bindParamPathIsNot(reg)

		// query array param path[isNot]
		if err := r.SetQueryParam("path[isNot]", joinedPathIsNot...); err != nil {
			return err
		}
	}

	if o.PathIs != nil {

		// binding items for path[is]
		joinedPathIs := o.bindParamPathIs(reg)

		// query array param path[is]
		if err := r.SetQueryParam("path[is]", joinedPathIs...); err != nil {
			return err
		}
	}

	if o.PathStart != nil {

		// query param path[start]
		var qrPathStart string

		if o.PathStart != nil {
			qrPathStart = *o.PathStart
		}
		qPathStart := qrPathStart
		if qPathStart != "" {

			if err := r.SetQueryParam("path[start]", qPathStart); err != nil {
				return err
			}
		}
	}

	if o.ProvidedPathIDIs != nil {

		// binding items for providedPathID[is]
		joinedProvidedPathIDIs := o.bindParamProvidedPathIDIs(reg)

		// query array param providedPathID[is]
		if err := r.SetQueryParam("providedPathID[is]", joinedProvidedPathIDIs...); err != nil {
			return err
		}
	}

	if o.ReconstructedPathIDIs != nil {

		// binding items for reconstructedPathID[is]
		joinedReconstructedPathIDIs := o.bindParamReconstructedPathIDIs(reg)

		// query array param reconstructedPathID[is]
		if err := r.SetQueryParam("reconstructedPathID[is]", joinedReconstructedPathIDIs...); err != nil {
			return err
		}
	}

	// query param showNonApi
	qrShowNonAPI := o.ShowNonAPI
	qShowNonAPI := swag.FormatBool(qrShowNonAPI)
	if qShowNonAPI != "" {

		if err := r.SetQueryParam("showNonApi", qShowNonAPI); err != nil {
			return err
		}
	}

	if o.SourceIPIsNot != nil {

		// binding items for sourceIP[isNot]
		joinedSourceIPIsNot := o.bindParamSourceIPIsNot(reg)

		// query array param sourceIP[isNot]
		if err := r.SetQueryParam("sourceIP[isNot]", joinedSourceIPIsNot...); err != nil {
			return err
		}
	}

	if o.SourceIPIs != nil {

		// binding items for sourceIP[is]
		joinedSourceIPIs := o.bindParamSourceIPIs(reg)

		// query array param sourceIP[is]
		if err := r.SetQueryParam("sourceIP[is]", joinedSourceIPIs...); err != nil {
			return err
		}
	}

	if o.SpecDiffTypeIs != nil {

		// binding items for specDiffType[is]
		joinedSpecDiffTypeIs := o.bindParamSpecDiffTypeIs(reg)

		// query array param specDiffType[is]
		if err := r.SetQueryParam("specDiffType[is]", joinedSpecDiffTypeIs...); err != nil {
			return err
		}
	}

	if o.SpecContains != nil {

		// binding items for spec[contains]
		joinedSpecContains := o.bindParamSpecContains(reg)

		// query array param spec[contains]
		if err := r.SetQueryParam("spec[contains]", joinedSpecContains...); err != nil {
			return err
		}
	}

	if o.SpecEnd != nil {

		// query param spec[end]
		var qrSpecEnd string

		if o.SpecEnd != nil {
			qrSpecEnd = *o.SpecEnd
		}
		qSpecEnd := qrSpecEnd
		if qSpecEnd != "" {

			if err := r.SetQueryParam("spec[end]", qSpecEnd); err != nil {
				return err
			}
		}
	}

	if o.SpecIsNot != nil {

		// binding items for spec[isNot]
		joinedSpecIsNot := o.bindParamSpecIsNot(reg)

		// query array param spec[isNot]
		if err := r.SetQueryParam("spec[isNot]", joinedSpecIsNot...); err != nil {
			return err
		}
	}

	if o.SpecIs != nil {

		// binding items for spec[is]
		joinedSpecIs := o.bindParamSpecIs(reg)

		// query array param spec[is]
		if err := r.SetQueryParam("spec[is]", joinedSpecIs...); err != nil {
			return err
		}
	}

	if o.SpecStart != nil {

		// query param spec[start]
		var qrSpecStart string

		if o.SpecStart != nil {
			qrSpecStart = *o.SpecStart
		}
		qSpecStart := qrSpecStart
		if qSpecStart != "" {

			if err := r.SetQueryParam("spec[start]", qSpecStart); err != nil {
				return err
			}
		}
	}

	// query param startTime
	qrStartTime := o.StartTime
	qStartTime := qrStartTime.String()
	if qStartTime != "" {

		if err := r.SetQueryParam("startTime", qStartTime); err != nil {
			return err
		}
	}

	if o.StatusCodeGte != nil {

		// query param statusCode[gte]
		var qrStatusCodeGte string

		if o.StatusCodeGte != nil {
			qrStatusCodeGte = *o.StatusCodeGte
		}
		qStatusCodeGte := qrStatusCodeGte
		if qStatusCodeGte != "" {

			if err := r.SetQueryParam("statusCode[gte]", qStatusCodeGte); err != nil {
				return err
			}
		}
	}

	if o.StatusCodeIsNot != nil {

		// binding items for statusCode[isNot]
		joinedStatusCodeIsNot := o.bindParamStatusCodeIsNot(reg)

		// query array param statusCode[isNot]
		if err := r.SetQueryParam("statusCode[isNot]", joinedStatusCodeIsNot...); err != nil {
			return err
		}
	}

	if o.StatusCodeIs != nil {

		// binding items for statusCode[is]
		joinedStatusCodeIs := o.bindParamStatusCodeIs(reg)

		// query array param statusCode[is]
		if err := r.SetQueryParam("statusCode[is]", joinedStatusCodeIs...); err != nil {
			return err
		}
	}

	if o.StatusCodeLte != nil {

		// query param statusCode[lte]
		var qrStatusCodeLte string

		if o.StatusCodeLte != nil {
			qrStatusCodeLte = *o.StatusCodeLte
		}
		qStatusCodeLte := qrStatusCodeLte
		if qStatusCodeLte != "" {

			if err := r.SetQueryParam("statusCode[lte]", qStatusCodeLte); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindParamGetAPIUsageHitCount binds the parameter destinationIP[isNot]
func (o *GetAPIUsageHitCountParams) bindParamDestinationIPIsNot(formats strfmt.Registry) []string {
	destinationIPIsNotIR := o.DestinationIPIsNot

	var destinationIPIsNotIC []string
	for _, destinationIPIsNotIIR := range destinationIPIsNotIR { // explode []string

		destinationIPIsNotIIV := destinationIPIsNotIIR // string as string
		destinationIPIsNotIC = append(destinationIPIsNotIC, destinationIPIsNotIIV)
	}

	// items.CollectionFormat: ""
	destinationIPIsNotIS := swag.JoinByFormat(destinationIPIsNotIC, "")

	return destinationIPIsNotIS
}

// bindParamGetAPIUsageHitCount binds the parameter destinationIP[is]
func (o *GetAPIUsageHitCountParams) bindParamDestinationIPIs(formats strfmt.Registry) []string {
	destinationIPIsIR := o.DestinationIPIs

	var destinationIPIsIC []string
	for _, destinationIPIsIIR := range destinationIPIsIR { // explode []string

		destinationIPIsIIV := destinationIPIsIIR // string as string
		destinationIPIsIC = append(destinationIPIsIC, destinationIPIsIIV)
	}

	// items.CollectionFormat: ""
	destinationIPIsIS := swag.JoinByFormat(destinationIPIsIC, "")

	return destinationIPIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter destinationPort[isNot]
func (o *GetAPIUsageHitCountParams) bindParamDestinationPortIsNot(formats strfmt.Registry) []string {
	destinationPortIsNotIR := o.DestinationPortIsNot

	var destinationPortIsNotIC []string
	for _, destinationPortIsNotIIR := range destinationPortIsNotIR { // explode []string

		destinationPortIsNotIIV := destinationPortIsNotIIR // string as string
		destinationPortIsNotIC = append(destinationPortIsNotIC, destinationPortIsNotIIV)
	}

	// items.CollectionFormat: ""
	destinationPortIsNotIS := swag.JoinByFormat(destinationPortIsNotIC, "")

	return destinationPortIsNotIS
}

// bindParamGetAPIUsageHitCount binds the parameter destinationPort[is]
func (o *GetAPIUsageHitCountParams) bindParamDestinationPortIs(formats strfmt.Registry) []string {
	destinationPortIsIR := o.DestinationPortIs

	var destinationPortIsIC []string
	for _, destinationPortIsIIR := range destinationPortIsIR { // explode []string

		destinationPortIsIIV := destinationPortIsIIR // string as string
		destinationPortIsIC = append(destinationPortIsIC, destinationPortIsIIV)
	}

	// items.CollectionFormat: ""
	destinationPortIsIS := swag.JoinByFormat(destinationPortIsIC, "")

	return destinationPortIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter method[is]
func (o *GetAPIUsageHitCountParams) bindParamMethodIs(formats strfmt.Registry) []string {
	methodIsIR := o.MethodIs

	var methodIsIC []string
	for _, methodIsIIR := range methodIsIR { // explode []string

		methodIsIIV := methodIsIIR // string as string
		methodIsIC = append(methodIsIC, methodIsIIV)
	}

	// items.CollectionFormat: ""
	methodIsIS := swag.JoinByFormat(methodIsIC, "")

	return methodIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter path[contains]
func (o *GetAPIUsageHitCountParams) bindParamPathContains(formats strfmt.Registry) []string {
	pathContainsIR := o.PathContains

	var pathContainsIC []string
	for _, pathContainsIIR := range pathContainsIR { // explode []string

		pathContainsIIV := pathContainsIIR // string as string
		pathContainsIC = append(pathContainsIC, pathContainsIIV)
	}

	// items.CollectionFormat: ""
	pathContainsIS := swag.JoinByFormat(pathContainsIC, "")

	return pathContainsIS
}

// bindParamGetAPIUsageHitCount binds the parameter path[isNot]
func (o *GetAPIUsageHitCountParams) bindParamPathIsNot(formats strfmt.Registry) []string {
	pathIsNotIR := o.PathIsNot

	var pathIsNotIC []string
	for _, pathIsNotIIR := range pathIsNotIR { // explode []string

		pathIsNotIIV := pathIsNotIIR // string as string
		pathIsNotIC = append(pathIsNotIC, pathIsNotIIV)
	}

	// items.CollectionFormat: ""
	pathIsNotIS := swag.JoinByFormat(pathIsNotIC, "")

	return pathIsNotIS
}

// bindParamGetAPIUsageHitCount binds the parameter path[is]
func (o *GetAPIUsageHitCountParams) bindParamPathIs(formats strfmt.Registry) []string {
	pathIsIR := o.PathIs

	var pathIsIC []string
	for _, pathIsIIR := range pathIsIR { // explode []string

		pathIsIIV := pathIsIIR // string as string
		pathIsIC = append(pathIsIC, pathIsIIV)
	}

	// items.CollectionFormat: ""
	pathIsIS := swag.JoinByFormat(pathIsIC, "")

	return pathIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter providedPathID[is]
func (o *GetAPIUsageHitCountParams) bindParamProvidedPathIDIs(formats strfmt.Registry) []string {
	providedPathIDIsIR := o.ProvidedPathIDIs

	var providedPathIDIsIC []string
	for _, providedPathIDIsIIR := range providedPathIDIsIR { // explode []string

		providedPathIDIsIIV := providedPathIDIsIIR // string as string
		providedPathIDIsIC = append(providedPathIDIsIC, providedPathIDIsIIV)
	}

	// items.CollectionFormat: ""
	providedPathIDIsIS := swag.JoinByFormat(providedPathIDIsIC, "")

	return providedPathIDIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter reconstructedPathID[is]
func (o *GetAPIUsageHitCountParams) bindParamReconstructedPathIDIs(formats strfmt.Registry) []string {
	reconstructedPathIDIsIR := o.ReconstructedPathIDIs

	var reconstructedPathIDIsIC []string
	for _, reconstructedPathIDIsIIR := range reconstructedPathIDIsIR { // explode []string

		reconstructedPathIDIsIIV := reconstructedPathIDIsIIR // string as string
		reconstructedPathIDIsIC = append(reconstructedPathIDIsIC, reconstructedPathIDIsIIV)
	}

	// items.CollectionFormat: ""
	reconstructedPathIDIsIS := swag.JoinByFormat(reconstructedPathIDIsIC, "")

	return reconstructedPathIDIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter sourceIP[isNot]
func (o *GetAPIUsageHitCountParams) bindParamSourceIPIsNot(formats strfmt.Registry) []string {
	sourceIPIsNotIR := o.SourceIPIsNot

	var sourceIPIsNotIC []string
	for _, sourceIPIsNotIIR := range sourceIPIsNotIR { // explode []string

		sourceIPIsNotIIV := sourceIPIsNotIIR // string as string
		sourceIPIsNotIC = append(sourceIPIsNotIC, sourceIPIsNotIIV)
	}

	// items.CollectionFormat: ""
	sourceIPIsNotIS := swag.JoinByFormat(sourceIPIsNotIC, "")

	return sourceIPIsNotIS
}

// bindParamGetAPIUsageHitCount binds the parameter sourceIP[is]
func (o *GetAPIUsageHitCountParams) bindParamSourceIPIs(formats strfmt.Registry) []string {
	sourceIPIsIR := o.SourceIPIs

	var sourceIPIsIC []string
	for _, sourceIPIsIIR := range sourceIPIsIR { // explode []string

		sourceIPIsIIV := sourceIPIsIIR // string as string
		sourceIPIsIC = append(sourceIPIsIC, sourceIPIsIIV)
	}

	// items.CollectionFormat: ""
	sourceIPIsIS := swag.JoinByFormat(sourceIPIsIC, "")

	return sourceIPIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter specDiffType[is]
func (o *GetAPIUsageHitCountParams) bindParamSpecDiffTypeIs(formats strfmt.Registry) []string {
	specDiffTypeIsIR := o.SpecDiffTypeIs

	var specDiffTypeIsIC []string
	for _, specDiffTypeIsIIR := range specDiffTypeIsIR { // explode []string

		specDiffTypeIsIIV := specDiffTypeIsIIR // string as string
		specDiffTypeIsIC = append(specDiffTypeIsIC, specDiffTypeIsIIV)
	}

	// items.CollectionFormat: ""
	specDiffTypeIsIS := swag.JoinByFormat(specDiffTypeIsIC, "")

	return specDiffTypeIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter spec[contains]
func (o *GetAPIUsageHitCountParams) bindParamSpecContains(formats strfmt.Registry) []string {
	specContainsIR := o.SpecContains

	var specContainsIC []string
	for _, specContainsIIR := range specContainsIR { // explode []string

		specContainsIIV := specContainsIIR // string as string
		specContainsIC = append(specContainsIC, specContainsIIV)
	}

	// items.CollectionFormat: ""
	specContainsIS := swag.JoinByFormat(specContainsIC, "")

	return specContainsIS
}

// bindParamGetAPIUsageHitCount binds the parameter spec[isNot]
func (o *GetAPIUsageHitCountParams) bindParamSpecIsNot(formats strfmt.Registry) []string {
	specIsNotIR := o.SpecIsNot

	var specIsNotIC []string
	for _, specIsNotIIR := range specIsNotIR { // explode []string

		specIsNotIIV := specIsNotIIR // string as string
		specIsNotIC = append(specIsNotIC, specIsNotIIV)
	}

	// items.CollectionFormat: ""
	specIsNotIS := swag.JoinByFormat(specIsNotIC, "")

	return specIsNotIS
}

// bindParamGetAPIUsageHitCount binds the parameter spec[is]
func (o *GetAPIUsageHitCountParams) bindParamSpecIs(formats strfmt.Registry) []string {
	specIsIR := o.SpecIs

	var specIsIC []string
	for _, specIsIIR := range specIsIR { // explode []string

		specIsIIV := specIsIIR // string as string
		specIsIC = append(specIsIC, specIsIIV)
	}

	// items.CollectionFormat: ""
	specIsIS := swag.JoinByFormat(specIsIC, "")

	return specIsIS
}

// bindParamGetAPIUsageHitCount binds the parameter statusCode[isNot]
func (o *GetAPIUsageHitCountParams) bindParamStatusCodeIsNot(formats strfmt.Registry) []string {
	statusCodeIsNotIR := o.StatusCodeIsNot

	var statusCodeIsNotIC []string
	for _, statusCodeIsNotIIR := range statusCodeIsNotIR { // explode []string

		statusCodeIsNotIIV := statusCodeIsNotIIR // string as string
		statusCodeIsNotIC = append(statusCodeIsNotIC, statusCodeIsNotIIV)
	}

	// items.CollectionFormat: ""
	statusCodeIsNotIS := swag.JoinByFormat(statusCodeIsNotIC, "")

	return statusCodeIsNotIS
}

// bindParamGetAPIUsageHitCount binds the parameter statusCode[is]
func (o *GetAPIUsageHitCountParams) bindParamStatusCodeIs(formats strfmt.Registry) []string {
	statusCodeIsIR := o.StatusCodeIs

	var statusCodeIsIC []string
	for _, statusCodeIsIIR := range statusCodeIsIR { // explode []string

		statusCodeIsIIV := statusCodeIsIIR // string as string
		statusCodeIsIC = append(statusCodeIsIC, statusCodeIsIIV)
	}

	// items.CollectionFormat: ""
	statusCodeIsIS := swag.JoinByFormat(statusCodeIsIC, "")

	return statusCodeIsIS
}
