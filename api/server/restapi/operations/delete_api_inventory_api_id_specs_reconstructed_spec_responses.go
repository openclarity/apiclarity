// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/apiclarity/apiclarity/api/server/models"
)

// DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreatedCode is the HTTP code returned for type DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated
const DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreatedCode int = 201

/*DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated Success

swagger:response deleteApiInventoryApiIdSpecsReconstructedSpecCreated
*/
type DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated struct {

	/*
	  In: Body
	*/
	Payload interface{} `json:"body,omitempty"`
}

// NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated creates DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated with default headers values
func NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated() *DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated {

	return &DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated{}
}

// WithPayload adds the payload to the delete Api inventory Api Id specs reconstructed spec created response
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated) WithPayload(payload interface{}) *DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete Api inventory Api Id specs reconstructed spec created response
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated) SetPayload(payload interface{}) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

/*DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault unknown error

swagger:response deleteApiInventoryApiIdSpecsReconstructedSpecDefault
*/
type DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault creates DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault with default headers values
func NewDeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault(code int) *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault {
	if code <= 0 {
		code = 500
	}

	return &DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the delete API inventory API ID specs reconstructed spec default response
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault) WithStatusCode(code int) *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the delete API inventory API ID specs reconstructed spec default response
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the delete API inventory API ID specs reconstructed spec default response
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault) WithPayload(payload *models.APIResponse) *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete API inventory API ID specs reconstructed spec default response
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteAPIInventoryAPIIDSpecsReconstructedSpecDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
