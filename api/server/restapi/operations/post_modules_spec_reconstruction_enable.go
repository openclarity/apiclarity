// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// PostModulesSpecReconstructionEnableHandlerFunc turns a function with the right signature into a post modules spec reconstruction enable handler
type PostModulesSpecReconstructionEnableHandlerFunc func(PostModulesSpecReconstructionEnableParams) middleware.Responder

// Handle executing the request and returning a response
func (fn PostModulesSpecReconstructionEnableHandlerFunc) Handle(params PostModulesSpecReconstructionEnableParams) middleware.Responder {
	return fn(params)
}

// PostModulesSpecReconstructionEnableHandler interface for that can handle valid post modules spec reconstruction enable params
type PostModulesSpecReconstructionEnableHandler interface {
	Handle(PostModulesSpecReconstructionEnableParams) middleware.Responder
}

// NewPostModulesSpecReconstructionEnable creates a new http.Handler for the post modules spec reconstruction enable operation
func NewPostModulesSpecReconstructionEnable(ctx *middleware.Context, handler PostModulesSpecReconstructionEnableHandler) *PostModulesSpecReconstructionEnable {
	return &PostModulesSpecReconstructionEnable{Context: ctx, Handler: handler}
}

/* PostModulesSpecReconstructionEnable swagger:route POST /modules/spec_reconstruction/enable postModulesSpecReconstructionEnable

enable/disable the spec reconstruction

enable/disable the spec reconstruction.

*/
type PostModulesSpecReconstructionEnable struct {
	Context *middleware.Context
	Handler PostModulesSpecReconstructionEnableHandler
}

func (o *PostModulesSpecReconstructionEnable) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewPostModulesSpecReconstructionEnableParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}