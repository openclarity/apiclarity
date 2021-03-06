// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetAPIInventoryAPIIDSuggestedReviewHandlerFunc turns a function with the right signature into a get API inventory API ID suggested review handler
type GetAPIInventoryAPIIDSuggestedReviewHandlerFunc func(GetAPIInventoryAPIIDSuggestedReviewParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAPIInventoryAPIIDSuggestedReviewHandlerFunc) Handle(params GetAPIInventoryAPIIDSuggestedReviewParams) middleware.Responder {
	return fn(params)
}

// GetAPIInventoryAPIIDSuggestedReviewHandler interface for that can handle valid get API inventory API ID suggested review params
type GetAPIInventoryAPIIDSuggestedReviewHandler interface {
	Handle(GetAPIInventoryAPIIDSuggestedReviewParams) middleware.Responder
}

// NewGetAPIInventoryAPIIDSuggestedReview creates a new http.Handler for the get API inventory API ID suggested review operation
func NewGetAPIInventoryAPIIDSuggestedReview(ctx *middleware.Context, handler GetAPIInventoryAPIIDSuggestedReviewHandler) *GetAPIInventoryAPIIDSuggestedReview {
	return &GetAPIInventoryAPIIDSuggestedReview{Context: ctx, Handler: handler}
}

/* GetAPIInventoryAPIIDSuggestedReview swagger:route GET /apiInventory/{apiId}/suggestedReview getApiInventoryApiIdSuggestedReview

Get reconstructed spec for review

*/
type GetAPIInventoryAPIIDSuggestedReview struct {
	Context *middleware.Context
	Handler GetAPIInventoryAPIIDSuggestedReviewHandler
}

func (o *GetAPIInventoryAPIIDSuggestedReview) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetAPIInventoryAPIIDSuggestedReviewParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
