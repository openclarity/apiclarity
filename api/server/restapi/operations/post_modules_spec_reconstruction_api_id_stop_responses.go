// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openclarity/apiclarity/api/server/models"
)

// PostModulesSpecReconstructionAPIIDStopNoContentCode is the HTTP code returned for type PostModulesSpecReconstructionAPIIDStopNoContent
const PostModulesSpecReconstructionAPIIDStopNoContentCode int = 204

/*PostModulesSpecReconstructionAPIIDStopNoContent Success

swagger:response postModulesSpecReconstructionApiIdStopNoContent
*/
type PostModulesSpecReconstructionAPIIDStopNoContent struct {
}

// NewPostModulesSpecReconstructionAPIIDStopNoContent creates PostModulesSpecReconstructionAPIIDStopNoContent with default headers values
func NewPostModulesSpecReconstructionAPIIDStopNoContent() *PostModulesSpecReconstructionAPIIDStopNoContent {

	return &PostModulesSpecReconstructionAPIIDStopNoContent{}
}

// WriteResponse to the client
func (o *PostModulesSpecReconstructionAPIIDStopNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

/*PostModulesSpecReconstructionAPIIDStopDefault unknown error

swagger:response postModulesSpecReconstructionApiIdStopDefault
*/
type PostModulesSpecReconstructionAPIIDStopDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewPostModulesSpecReconstructionAPIIDStopDefault creates PostModulesSpecReconstructionAPIIDStopDefault with default headers values
func NewPostModulesSpecReconstructionAPIIDStopDefault(code int) *PostModulesSpecReconstructionAPIIDStopDefault {
	if code <= 0 {
		code = 500
	}

	return &PostModulesSpecReconstructionAPIIDStopDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the post modules spec reconstruction API ID stop default response
func (o *PostModulesSpecReconstructionAPIIDStopDefault) WithStatusCode(code int) *PostModulesSpecReconstructionAPIIDStopDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the post modules spec reconstruction API ID stop default response
func (o *PostModulesSpecReconstructionAPIIDStopDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the post modules spec reconstruction API ID stop default response
func (o *PostModulesSpecReconstructionAPIIDStopDefault) WithPayload(payload *models.APIResponse) *PostModulesSpecReconstructionAPIIDStopDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post modules spec reconstruction API ID stop default response
func (o *PostModulesSpecReconstructionAPIIDStopDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostModulesSpecReconstructionAPIIDStopDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
