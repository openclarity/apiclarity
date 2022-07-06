// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// GetFeaturesHandlerFunc turns a function with the right signature into a get features handler
type GetFeaturesHandlerFunc func(GetFeaturesParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetFeaturesHandlerFunc) Handle(params GetFeaturesParams) middleware.Responder {
	return fn(params)
}

// GetFeaturesHandler interface for that can handle valid get features params
type GetFeaturesHandler interface {
	Handle(GetFeaturesParams) middleware.Responder
}

// NewGetFeatures creates a new http.Handler for the get features operation
func NewGetFeatures(ctx *middleware.Context, handler GetFeaturesHandler) *GetFeatures {
	return &GetFeatures{Context: ctx, Handler: handler}
}

/* GetFeatures swagger:route GET /features getFeatures

Get the list of APIClarity features and for each feature the API hosts the feature requires to get trace for

*/
type GetFeatures struct {
	Context *middleware.Context
	Handler GetFeaturesHandler
}

func (o *GetFeatures) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewGetFeaturesParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
