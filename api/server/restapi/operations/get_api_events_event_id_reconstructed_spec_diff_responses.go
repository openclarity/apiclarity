// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openclarity/apiclarity/api/server/models"
)

// GetAPIEventsEventIDReconstructedSpecDiffOKCode is the HTTP code returned for type GetAPIEventsEventIDReconstructedSpecDiffOK
const GetAPIEventsEventIDReconstructedSpecDiffOKCode int = 200

/*GetAPIEventsEventIDReconstructedSpecDiffOK Success

swagger:response getApiEventsEventIdReconstructedSpecDiffOK
*/
type GetAPIEventsEventIDReconstructedSpecDiffOK struct {

	/*
	  In: Body
	*/
	Payload *models.APIEventSpecDiff `json:"body,omitempty"`
}

// NewGetAPIEventsEventIDReconstructedSpecDiffOK creates GetAPIEventsEventIDReconstructedSpecDiffOK with default headers values
func NewGetAPIEventsEventIDReconstructedSpecDiffOK() *GetAPIEventsEventIDReconstructedSpecDiffOK {

	return &GetAPIEventsEventIDReconstructedSpecDiffOK{}
}

// WithPayload adds the payload to the get Api events event Id reconstructed spec diff o k response
func (o *GetAPIEventsEventIDReconstructedSpecDiffOK) WithPayload(payload *models.APIEventSpecDiff) *GetAPIEventsEventIDReconstructedSpecDiffOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get Api events event Id reconstructed spec diff o k response
func (o *GetAPIEventsEventIDReconstructedSpecDiffOK) SetPayload(payload *models.APIEventSpecDiff) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAPIEventsEventIDReconstructedSpecDiffOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*GetAPIEventsEventIDReconstructedSpecDiffDefault unknown error

swagger:response getApiEventsEventIdReconstructedSpecDiffDefault
*/
type GetAPIEventsEventIDReconstructedSpecDiffDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewGetAPIEventsEventIDReconstructedSpecDiffDefault creates GetAPIEventsEventIDReconstructedSpecDiffDefault with default headers values
func NewGetAPIEventsEventIDReconstructedSpecDiffDefault(code int) *GetAPIEventsEventIDReconstructedSpecDiffDefault {
	if code <= 0 {
		code = 500
	}

	return &GetAPIEventsEventIDReconstructedSpecDiffDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the get API events event ID reconstructed spec diff default response
func (o *GetAPIEventsEventIDReconstructedSpecDiffDefault) WithStatusCode(code int) *GetAPIEventsEventIDReconstructedSpecDiffDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the get API events event ID reconstructed spec diff default response
func (o *GetAPIEventsEventIDReconstructedSpecDiffDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the get API events event ID reconstructed spec diff default response
func (o *GetAPIEventsEventIDReconstructedSpecDiffDefault) WithPayload(payload *models.APIResponse) *GetAPIEventsEventIDReconstructedSpecDiffDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get API events event ID reconstructed spec diff default response
func (o *GetAPIEventsEventIDReconstructedSpecDiffDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAPIEventsEventIDReconstructedSpecDiffDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
