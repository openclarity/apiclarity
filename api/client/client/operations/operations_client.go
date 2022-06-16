// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new operations API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	DeleteAPIInventoryAPIIDSpecsProvidedSpec(params *DeleteAPIInventoryAPIIDSpecsProvidedSpecParams, opts ...ClientOption) (*DeleteAPIInventoryAPIIDSpecsProvidedSpecOK, error)

	DeleteAPIInventoryAPIIDSpecsReconstructedSpec(params *DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams, opts ...ClientOption) (*DeleteAPIInventoryAPIIDSpecsReconstructedSpecOK, error)

	GetAPIEvents(params *GetAPIEventsParams, opts ...ClientOption) (*GetAPIEventsOK, error)

	GetAPIEventsEventID(params *GetAPIEventsEventIDParams, opts ...ClientOption) (*GetAPIEventsEventIDOK, error)

	GetAPIEventsEventIDProvidedSpecDiff(params *GetAPIEventsEventIDProvidedSpecDiffParams, opts ...ClientOption) (*GetAPIEventsEventIDProvidedSpecDiffOK, error)

	GetAPIEventsEventIDReconstructedSpecDiff(params *GetAPIEventsEventIDReconstructedSpecDiffParams, opts ...ClientOption) (*GetAPIEventsEventIDReconstructedSpecDiffOK, error)

	GetAPIInventory(params *GetAPIInventoryParams, opts ...ClientOption) (*GetAPIInventoryOK, error)

	GetAPIInventoryAPIIDAPIInfo(params *GetAPIInventoryAPIIDAPIInfoParams, opts ...ClientOption) (*GetAPIInventoryAPIIDAPIInfoOK, error)

	GetAPIInventoryAPIIDFromHostAndPort(params *GetAPIInventoryAPIIDFromHostAndPortParams, opts ...ClientOption) (*GetAPIInventoryAPIIDFromHostAndPortOK, error)

	GetAPIInventoryAPIIDProvidedSwaggerJSON(params *GetAPIInventoryAPIIDProvidedSwaggerJSONParams, opts ...ClientOption) (*GetAPIInventoryAPIIDProvidedSwaggerJSONOK, error)

	GetAPIInventoryAPIIDReconstructedSwaggerJSON(params *GetAPIInventoryAPIIDReconstructedSwaggerJSONParams, opts ...ClientOption) (*GetAPIInventoryAPIIDReconstructedSwaggerJSONOK, error)

	GetAPIInventoryAPIIDSpecs(params *GetAPIInventoryAPIIDSpecsParams, opts ...ClientOption) (*GetAPIInventoryAPIIDSpecsOK, error)

	GetAPIInventoryAPIIDSuggestedReview(params *GetAPIInventoryAPIIDSuggestedReviewParams, opts ...ClientOption) (*GetAPIInventoryAPIIDSuggestedReviewOK, error)

	GetAPIUsageHitCount(params *GetAPIUsageHitCountParams, opts ...ClientOption) (*GetAPIUsageHitCountOK, error)

	GetDashboardAPIUsage(params *GetDashboardAPIUsageParams, opts ...ClientOption) (*GetDashboardAPIUsageOK, error)

	GetDashboardAPIUsageLatestDiffs(params *GetDashboardAPIUsageLatestDiffsParams, opts ...ClientOption) (*GetDashboardAPIUsageLatestDiffsOK, error)

	GetDashboardAPIUsageMostUsed(params *GetDashboardAPIUsageMostUsedParams, opts ...ClientOption) (*GetDashboardAPIUsageMostUsedOK, error)

	PostAPIInventory(params *PostAPIInventoryParams, opts ...ClientOption) (*PostAPIInventoryOK, error)

	PostAPIInventoryReviewIDApprovedReview(params *PostAPIInventoryReviewIDApprovedReviewParams, opts ...ClientOption) (*PostAPIInventoryReviewIDApprovedReviewOK, error)

	PutAPIInventoryAPIIDSpecsProvidedSpec(params *PutAPIInventoryAPIIDSpecsProvidedSpecParams, opts ...ClientOption) (*PutAPIInventoryAPIIDSpecsProvidedSpecCreated, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  DeleteAPIInventoryAPIIDSpecsProvidedSpec unsets a provided spec for a specific API
*/
func (a *Client) DeleteAPIInventoryAPIIDSpecsProvidedSpec(params *DeleteAPIInventoryAPIIDSpecsProvidedSpecParams, opts ...ClientOption) (*DeleteAPIInventoryAPIIDSpecsProvidedSpecOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteAPIInventoryAPIIDSpecsProvidedSpecParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "DeleteAPIInventoryAPIIDSpecsProvidedSpec",
		Method:             "DELETE",
		PathPattern:        "/apiInventory/{apiId}/specs/providedSpec",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteAPIInventoryAPIIDSpecsProvidedSpecReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteAPIInventoryAPIIDSpecsProvidedSpecOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  DeleteAPIInventoryAPIIDSpecsReconstructedSpec unsets a reconstructed spec for a specific API
*/
func (a *Client) DeleteAPIInventoryAPIIDSpecsReconstructedSpec(params *DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams, opts ...ClientOption) (*DeleteAPIInventoryAPIIDSpecsReconstructedSpecOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "DeleteAPIInventoryAPIIDSpecsReconstructedSpec",
		Method:             "DELETE",
		PathPattern:        "/apiInventory/{apiId}/specs/reconstructedSpec",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &DeleteAPIInventoryAPIIDSpecsReconstructedSpecReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteAPIInventoryAPIIDSpecsReconstructedSpecOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIEvents gets API events
*/
func (a *Client) GetAPIEvents(params *GetAPIEventsParams, opts ...ClientOption) (*GetAPIEventsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIEventsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIEvents",
		Method:             "GET",
		PathPattern:        "/apiEvents",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIEventsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIEventsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIEventsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIEventsEventID gets API event
*/
func (a *Client) GetAPIEventsEventID(params *GetAPIEventsEventIDParams, opts ...ClientOption) (*GetAPIEventsEventIDOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIEventsEventIDParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIEventsEventID",
		Method:             "GET",
		PathPattern:        "/apiEvents/{eventId}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIEventsEventIDReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIEventsEventIDOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIEventsEventIDDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIEventsEventIDProvidedSpecDiff gets API event provided spec diff
*/
func (a *Client) GetAPIEventsEventIDProvidedSpecDiff(params *GetAPIEventsEventIDProvidedSpecDiffParams, opts ...ClientOption) (*GetAPIEventsEventIDProvidedSpecDiffOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIEventsEventIDProvidedSpecDiffParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIEventsEventIDProvidedSpecDiff",
		Method:             "GET",
		PathPattern:        "/apiEvents/{eventId}/providedSpecDiff",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIEventsEventIDProvidedSpecDiffReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIEventsEventIDProvidedSpecDiffOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIEventsEventIDProvidedSpecDiffDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIEventsEventIDReconstructedSpecDiff gets API event reconstructed spec diff
*/
func (a *Client) GetAPIEventsEventIDReconstructedSpecDiff(params *GetAPIEventsEventIDReconstructedSpecDiffParams, opts ...ClientOption) (*GetAPIEventsEventIDReconstructedSpecDiffOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIEventsEventIDReconstructedSpecDiffParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIEventsEventIDReconstructedSpecDiff",
		Method:             "GET",
		PathPattern:        "/apiEvents/{eventId}/reconstructedSpecDiff",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIEventsEventIDReconstructedSpecDiffReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIEventsEventIDReconstructedSpecDiffOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIEventsEventIDReconstructedSpecDiffDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventory gets API inventory
*/
func (a *Client) GetAPIInventory(params *GetAPIInventoryParams, opts ...ClientOption) (*GetAPIInventoryOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventory",
		Method:             "GET",
		PathPattern:        "/apiInventory",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventoryAPIIDAPIInfo gets api info from api id
*/
func (a *Client) GetAPIInventoryAPIIDAPIInfo(params *GetAPIInventoryAPIIDAPIInfoParams, opts ...ClientOption) (*GetAPIInventoryAPIIDAPIInfoOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryAPIIDAPIInfoParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventoryAPIIDAPIInfo",
		Method:             "GET",
		PathPattern:        "/apiInventory/{apiId}/apiInfo",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryAPIIDAPIInfoReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryAPIIDAPIInfoOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryAPIIDAPIInfoDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventoryAPIIDFromHostAndPort gets api Id from host and port
*/
func (a *Client) GetAPIInventoryAPIIDFromHostAndPort(params *GetAPIInventoryAPIIDFromHostAndPortParams, opts ...ClientOption) (*GetAPIInventoryAPIIDFromHostAndPortOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryAPIIDFromHostAndPortParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventoryAPIIDFromHostAndPort",
		Method:             "GET",
		PathPattern:        "/apiInventory/apiId/fromHostAndPort",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryAPIIDFromHostAndPortReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryAPIIDFromHostAndPortOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryAPIIDFromHostAndPortDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventoryAPIIDProvidedSwaggerJSON gets provided API spec json file
*/
func (a *Client) GetAPIInventoryAPIIDProvidedSwaggerJSON(params *GetAPIInventoryAPIIDProvidedSwaggerJSONParams, opts ...ClientOption) (*GetAPIInventoryAPIIDProvidedSwaggerJSONOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryAPIIDProvidedSwaggerJSONParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventoryAPIIDProvidedSwaggerJSON",
		Method:             "GET",
		PathPattern:        "/apiInventory/{apiId}/provided_swagger.json",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryAPIIDProvidedSwaggerJSONReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryAPIIDProvidedSwaggerJSONOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryAPIIDProvidedSwaggerJSONDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventoryAPIIDReconstructedSwaggerJSON gets reconstructed API spec json file
*/
func (a *Client) GetAPIInventoryAPIIDReconstructedSwaggerJSON(params *GetAPIInventoryAPIIDReconstructedSwaggerJSONParams, opts ...ClientOption) (*GetAPIInventoryAPIIDReconstructedSwaggerJSONOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryAPIIDReconstructedSwaggerJSONParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventoryAPIIDReconstructedSwaggerJSON",
		Method:             "GET",
		PathPattern:        "/apiInventory/{apiId}/reconstructed_swagger.json",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryAPIIDReconstructedSwaggerJSONReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryAPIIDReconstructedSwaggerJSONOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryAPIIDReconstructedSwaggerJSONDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventoryAPIIDSpecs gets provided and reconstructed open api specs for a specific API
*/
func (a *Client) GetAPIInventoryAPIIDSpecs(params *GetAPIInventoryAPIIDSpecsParams, opts ...ClientOption) (*GetAPIInventoryAPIIDSpecsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryAPIIDSpecsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventoryAPIIDSpecs",
		Method:             "GET",
		PathPattern:        "/apiInventory/{apiId}/specs",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryAPIIDSpecsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryAPIIDSpecsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryAPIIDSpecsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIInventoryAPIIDSuggestedReview gets reconstructed spec for review
*/
func (a *Client) GetAPIInventoryAPIIDSuggestedReview(params *GetAPIInventoryAPIIDSuggestedReviewParams, opts ...ClientOption) (*GetAPIInventoryAPIIDSuggestedReviewOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIInventoryAPIIDSuggestedReviewParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIInventoryAPIIDSuggestedReview",
		Method:             "GET",
		PathPattern:        "/apiInventory/{apiId}/suggestedReview",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIInventoryAPIIDSuggestedReviewReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIInventoryAPIIDSuggestedReviewOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIInventoryAPIIDSuggestedReviewDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetAPIUsageHitCount gets a hit count within a selected timeframe for the filtered API events
*/
func (a *Client) GetAPIUsageHitCount(params *GetAPIUsageHitCountParams, opts ...ClientOption) (*GetAPIUsageHitCountOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAPIUsageHitCountParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAPIUsageHitCount",
		Method:             "GET",
		PathPattern:        "/apiUsage/hitCount",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetAPIUsageHitCountReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAPIUsageHitCountOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetAPIUsageHitCountDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetDashboardAPIUsage gets API usage
*/
func (a *Client) GetDashboardAPIUsage(params *GetDashboardAPIUsageParams, opts ...ClientOption) (*GetDashboardAPIUsageOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetDashboardAPIUsageParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetDashboardAPIUsage",
		Method:             "GET",
		PathPattern:        "/dashboard/apiUsage",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetDashboardAPIUsageReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetDashboardAPIUsageOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetDashboardAPIUsageDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetDashboardAPIUsageLatestDiffs gets latest spec diffs
*/
func (a *Client) GetDashboardAPIUsageLatestDiffs(params *GetDashboardAPIUsageLatestDiffsParams, opts ...ClientOption) (*GetDashboardAPIUsageLatestDiffsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetDashboardAPIUsageLatestDiffsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetDashboardAPIUsageLatestDiffs",
		Method:             "GET",
		PathPattern:        "/dashboard/apiUsage/latestDiffs",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetDashboardAPIUsageLatestDiffsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetDashboardAPIUsageLatestDiffsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetDashboardAPIUsageLatestDiffsDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  GetDashboardAPIUsageMostUsed gets most used a p is
*/
func (a *Client) GetDashboardAPIUsageMostUsed(params *GetDashboardAPIUsageMostUsedParams, opts ...ClientOption) (*GetDashboardAPIUsageMostUsedOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetDashboardAPIUsageMostUsedParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetDashboardAPIUsageMostUsed",
		Method:             "GET",
		PathPattern:        "/dashboard/apiUsage/mostUsed",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &GetDashboardAPIUsageMostUsedReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetDashboardAPIUsageMostUsedOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*GetDashboardAPIUsageMostUsedDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  PostAPIInventory creates API inventory item
*/
func (a *Client) PostAPIInventory(params *PostAPIInventoryParams, opts ...ClientOption) (*PostAPIInventoryOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostAPIInventoryParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PostAPIInventory",
		Method:             "POST",
		PathPattern:        "/apiInventory",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostAPIInventoryReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostAPIInventoryOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostAPIInventoryDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  PostAPIInventoryReviewIDApprovedReview applies the approved review to create the reconstructed spec
*/
func (a *Client) PostAPIInventoryReviewIDApprovedReview(params *PostAPIInventoryReviewIDApprovedReviewParams, opts ...ClientOption) (*PostAPIInventoryReviewIDApprovedReviewOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostAPIInventoryReviewIDApprovedReviewParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PostAPIInventoryReviewIDApprovedReview",
		Method:             "POST",
		PathPattern:        "/apiInventory/{reviewId}/approvedReview",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PostAPIInventoryReviewIDApprovedReviewReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostAPIInventoryReviewIDApprovedReviewOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PostAPIInventoryReviewIDApprovedReviewDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

/*
  PutAPIInventoryAPIIDSpecsProvidedSpec adds or edit a spec for a specific API
*/
func (a *Client) PutAPIInventoryAPIIDSpecsProvidedSpec(params *PutAPIInventoryAPIIDSpecsProvidedSpecParams, opts ...ClientOption) (*PutAPIInventoryAPIIDSpecsProvidedSpecCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPutAPIInventoryAPIIDSpecsProvidedSpecParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PutAPIInventoryAPIIDSpecsProvidedSpec",
		Method:             "PUT",
		PathPattern:        "/apiInventory/{apiId}/specs/providedSpec",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http"},
		Params:             params,
		Reader:             &PutAPIInventoryAPIIDSpecsProvidedSpecReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PutAPIInventoryAPIIDSpecsProvidedSpecCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	unexpectedSuccess := result.(*PutAPIInventoryAPIIDSpecsProvidedSpecDefault)
	return nil, runtime.NewAPIError("unexpected success response: content available as default response in error", unexpectedSuccess, unexpectedSuccess.Code())
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
