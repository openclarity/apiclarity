// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/openclarity/apiclarity/api/server/models"
)

// PostModulesSpecReconstructionAPIIDStartNoContentCode is the HTTP code returned for type PostModulesSpecReconstructionAPIIDStartNoContent
const PostModulesSpecReconstructionAPIIDStartNoContentCode int = 204

/*PostModulesSpecReconstructionAPIIDStartNoContent Success

swagger:response postModulesSpecReconstructionApiIdStartNoContent
*/
type PostModulesSpecReconstructionAPIIDStartNoContent struct {
}

// NewPostModulesSpecReconstructionAPIIDStartNoContent creates PostModulesSpecReconstructionAPIIDStartNoContent with default headers values
func NewPostModulesSpecReconstructionAPIIDStartNoContent() *PostModulesSpecReconstructionAPIIDStartNoContent {

	return &PostModulesSpecReconstructionAPIIDStartNoContent{}
}

// WriteResponse to the client
func (o *PostModulesSpecReconstructionAPIIDStartNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

/*PostModulesSpecReconstructionAPIIDStartDefault unknown error

swagger:response postModulesSpecReconstructionApiIdStartDefault
*/
type PostModulesSpecReconstructionAPIIDStartDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.APIResponse `json:"body,omitempty"`
}

// NewPostModulesSpecReconstructionAPIIDStartDefault creates PostModulesSpecReconstructionAPIIDStartDefault with default headers values
func NewPostModulesSpecReconstructionAPIIDStartDefault(code int) *PostModulesSpecReconstructionAPIIDStartDefault {
	if code <= 0 {
		code = 500
	}

	return &PostModulesSpecReconstructionAPIIDStartDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the post modules spec reconstruction API ID start default response
func (o *PostModulesSpecReconstructionAPIIDStartDefault) WithStatusCode(code int) *PostModulesSpecReconstructionAPIIDStartDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the post modules spec reconstruction API ID start default response
func (o *PostModulesSpecReconstructionAPIIDStartDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the post modules spec reconstruction API ID start default response
func (o *PostModulesSpecReconstructionAPIIDStartDefault) WithPayload(payload *models.APIResponse) *PostModulesSpecReconstructionAPIIDStartDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the post modules spec reconstruction API ID start default response
func (o *PostModulesSpecReconstructionAPIIDStartDefault) SetPayload(payload *models.APIResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *PostModulesSpecReconstructionAPIIDStartDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}