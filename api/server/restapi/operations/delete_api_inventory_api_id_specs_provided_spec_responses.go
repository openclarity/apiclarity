// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/apiclarity/apiclarity/api/server/models"
)

// DeleteAPIInventoryAPIIDSpecsProvidedSpecCreatedCode is the HTTP code returned for type DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated
const DeleteAPIInventoryAPIIDSpecsProvidedSpecCreatedCode int = 201

/*DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated Success

swagger:response deleteApiInventoryApiIdSpecsProvidedSpecCreated
*/
type DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated struct {

	/*
	  In: Body
	*/
	Payload interface{} `json:"body,omitempty"`
}

// NewDeleteAPIInventoryAPIIDSpecsProvidedSpecCreated creates DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated with default headers values
func NewDeleteAPIInventoryAPIIDSpecsProvidedSpecCreated() *DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated {

	return &DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated{}
}

// WithPayload adds the payload to the delete Api inventory Api Id specs provided spec created response
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated) WithPayload(payload interface{}) *DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete Api inventory Api Id specs provided spec created response
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated) SetPayload(payload interface{}) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

/*DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault unknown error

swagger:response deleteApiInventoryApiIdSpecsProvidedSpecDefault
*/
type DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewDeleteAPIInventoryAPIIDSpecsProvidedSpecDefault creates DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault with default headers values
func NewDeleteAPIInventoryAPIIDSpecsProvidedSpecDefault(code int) *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault {
	if code <= 0 {
		code = 500
	}

	return &DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the delete API inventory API ID specs provided spec default response
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault) WithStatusCode(code int) *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the delete API inventory API ID specs provided spec default response
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the delete API inventory API ID specs provided spec default response
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault) WithPayload(payload *models.APIResponse) *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete API inventory API ID specs provided spec default response
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteAPIInventoryAPIIDSpecsProvidedSpecDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
