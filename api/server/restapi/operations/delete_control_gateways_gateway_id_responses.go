// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openclarity/apiclarity/api/server/models"
)

// DeleteControlGatewaysGatewayIDNoContentCode is the HTTP code returned for type DeleteControlGatewaysGatewayIDNoContent
const DeleteControlGatewaysGatewayIDNoContentCode int = 204

/*DeleteControlGatewaysGatewayIDNoContent Success

swagger:response deleteControlGatewaysGatewayIdNoContent
*/
type DeleteControlGatewaysGatewayIDNoContent struct {
}

// NewDeleteControlGatewaysGatewayIDNoContent creates DeleteControlGatewaysGatewayIDNoContent with default headers values
func NewDeleteControlGatewaysGatewayIDNoContent() *DeleteControlGatewaysGatewayIDNoContent {

	return &DeleteControlGatewaysGatewayIDNoContent{}
}

// WriteResponse to the client
func (o *DeleteControlGatewaysGatewayIDNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

// DeleteControlGatewaysGatewayIDNotFoundCode is the HTTP code returned for type DeleteControlGatewaysGatewayIDNotFound
const DeleteControlGatewaysGatewayIDNotFoundCode int = 404

/*DeleteControlGatewaysGatewayIDNotFound API Gateway not found

swagger:response deleteControlGatewaysGatewayIdNotFound
*/
type DeleteControlGatewaysGatewayIDNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewDeleteControlGatewaysGatewayIDNotFound creates DeleteControlGatewaysGatewayIDNotFound with default headers values
func NewDeleteControlGatewaysGatewayIDNotFound() *DeleteControlGatewaysGatewayIDNotFound {

	return &DeleteControlGatewaysGatewayIDNotFound{}
}

// WithPayload adds the payload to the delete control gateways gateway Id not found response
func (o *DeleteControlGatewaysGatewayIDNotFound) WithPayload(payload *models.APIResponse) *DeleteControlGatewaysGatewayIDNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete control gateways gateway Id not found response
func (o *DeleteControlGatewaysGatewayIDNotFound) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteControlGatewaysGatewayIDNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*DeleteControlGatewaysGatewayIDDefault unknown error

swagger:response deleteControlGatewaysGatewayIdDefault
*/
type DeleteControlGatewaysGatewayIDDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewDeleteControlGatewaysGatewayIDDefault creates DeleteControlGatewaysGatewayIDDefault with default headers values
func NewDeleteControlGatewaysGatewayIDDefault(code int) *DeleteControlGatewaysGatewayIDDefault {
	if code <= 0 {
		code = 500
	}

	return &DeleteControlGatewaysGatewayIDDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the delete control gateways gateway ID default response
func (o *DeleteControlGatewaysGatewayIDDefault) WithStatusCode(code int) *DeleteControlGatewaysGatewayIDDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the delete control gateways gateway ID default response
func (o *DeleteControlGatewaysGatewayIDDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the delete control gateways gateway ID default response
func (o *DeleteControlGatewaysGatewayIDDefault) WithPayload(payload *models.APIResponse) *DeleteControlGatewaysGatewayIDDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete control gateways gateway ID default response
func (o *DeleteControlGatewaysGatewayIDDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteControlGatewaysGatewayIDDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
