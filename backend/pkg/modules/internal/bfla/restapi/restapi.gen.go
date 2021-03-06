// Package restapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.1 DO NOT EDIT.
package restapi

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// Defines values for BFLAStatus.
const (
	BFLAStatusLEARNING BFLAStatus = "LEARNING"

	BFLAStatusLEGITIMATE BFLAStatus = "LEGITIMATE"

	BFLAStatusNOSPEC BFLAStatus = "NO_SPEC"

	BFLAStatusSUSPICIOUSHIGH BFLAStatus = "SUSPICIOUS_HIGH"

	BFLAStatusSUSPICIOUSMEDIUM BFLAStatus = "SUSPICIOUS_MEDIUM"
)

// Defines values for DetectedUserSource.
const (
	DetectedUserSourceBASIC DetectedUserSource = "BASIC"

	DetectedUserSourceJWT DetectedUserSource = "JWT"

	DetectedUserSourceKONGXCONSUMERID DetectedUserSource = "KONG_X_CONSUMER_ID"
)

// Defines values for OperationEnum.
const (
	OperationEnumApprove OperationEnum = "approve"

	OperationEnumApproveUser OperationEnum = "approve_user"

	OperationEnumDeny OperationEnum = "deny"

	OperationEnumDenyUser OperationEnum = "deny_user"
)

// Defines values for SpecType.
const (
	SpecTypeNONE SpecType = "NONE"

	SpecTypePROVIDED SpecType = "PROVIDED"

	SpecTypeRECONSTRUCTED SpecType = "RECONSTRUCTED"
)

// APIEventAnnotations defines model for APIEventAnnotations.
type APIEventAnnotations struct {
	BflaStatus           BFLAStatus    `json:"bflaStatus"`
	DestinationK8sObject *K8sObjectRef `json:"destinationK8sObject,omitempty"`
	DetectedUser         *DetectedUser `json:"detectedUser,omitempty"`
	External             bool          `json:"external"`
	SourceK8sObject      *K8sObjectRef `json:"sourceK8sObject,omitempty"`
}

// An object that is return in all cases of failures.
type ApiResponse struct {
	Message string `json:"message"`
}

// AuthorizationModel defines model for AuthorizationModel.
type AuthorizationModel struct {
	Learning   bool                          `json:"learning"`
	Operations []AuthorizationModelOperation `json:"operations"`
	SpecType   SpecType                      `json:"specType"`
}

// AuthorizationModelAudience defines model for AuthorizationModelAudience.
type AuthorizationModelAudience struct {
	Authorized bool           `json:"authorized"`
	EndUsers   []DetectedUser `json:"end_users"`
	External   bool           `json:"external"`
	K8sObject  *K8sObjectRef  `json:"k8s_object,omitempty"`
}

// AuthorizationModelOperation defines model for AuthorizationModelOperation.
type AuthorizationModelOperation struct {
	Audience []AuthorizationModelAudience `json:"audience"`
	Method   string                       `json:"method"`
	Path     string                       `json:"path"`
}

// BFLAStatus defines model for BFLAStatus.
type BFLAStatus string

// DetectedUser defines model for DetectedUser.
type DetectedUser struct {
	Id        string             `json:"id"`
	IpAddress string             `json:"ip_address"`
	Source    DetectedUserSource `json:"source"`
}

// DetectedUserSource defines model for DetectedUser.Source.
type DetectedUserSource string

// K8sObjectRef defines model for K8sObjectRef.
type K8sObjectRef struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Uid        string `json:"uid"`
}

// OperationEnum defines model for OperationEnum.
type OperationEnum string

// SpecType defines model for SpecType.
type SpecType string

// Version defines model for Version.
type Version struct {
	Version string `json:"version"`
}

// PutAuthorizationModelApiIDApproveParams defines parameters for PutAuthorizationModelApiIDApprove.
type PutAuthorizationModelApiIDApproveParams struct {
	Method       string `json:"method"`
	Path         string `json:"path"`
	K8sClientUid string `json:"k8sClientUid"`
}

// PutAuthorizationModelApiIDDenyParams defines parameters for PutAuthorizationModelApiIDDeny.
type PutAuthorizationModelApiIDDenyParams struct {
	Method       string `json:"method"`
	Path         string `json:"path"`
	K8sClientUid string `json:"k8sClientUid"`
}

// PutAuthorizationModelApiIDLearningResetParams defines parameters for PutAuthorizationModelApiIDLearningReset.
type PutAuthorizationModelApiIDLearningResetParams struct {
	NrTraces int `json:"nr_traces"`
}

// PutAuthorizationModelApiIDLearningStartParams defines parameters for PutAuthorizationModelApiIDLearningStart.
type PutAuthorizationModelApiIDLearningStartParams struct {
	NrTraces int `json:"nr_traces"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {

	// (GET /authorizationModel/{apiID})
	GetAuthorizationModelApiID(w http.ResponseWriter, r *http.Request, apiID int)

	// (PUT /authorizationModel/{apiID}/approve)
	PutAuthorizationModelApiIDApprove(w http.ResponseWriter, r *http.Request, apiID int, params PutAuthorizationModelApiIDApproveParams)

	// (PUT /authorizationModel/{apiID}/deny)
	PutAuthorizationModelApiIDDeny(w http.ResponseWriter, r *http.Request, apiID int, params PutAuthorizationModelApiIDDenyParams)

	// (PUT /authorizationModel/{apiID}/learning/reset)
	PutAuthorizationModelApiIDLearningReset(w http.ResponseWriter, r *http.Request, apiID int, params PutAuthorizationModelApiIDLearningResetParams)

	// (PUT /authorizationModel/{apiID}/learning/start)
	PutAuthorizationModelApiIDLearningStart(w http.ResponseWriter, r *http.Request, apiID int, params PutAuthorizationModelApiIDLearningStartParams)

	// (PUT /authorizationModel/{apiID}/learning/stop)
	PutAuthorizationModelApiIDLearningStop(w http.ResponseWriter, r *http.Request, apiID int)
	// Get the event with the annotations and bfla status
	// (GET /event/{id})
	GetEvent(w http.ResponseWriter, r *http.Request, id int)

	// (PUT /event/{id}/{operation})
	PutEventIdOperation(w http.ResponseWriter, r *http.Request, id int, operation OperationEnum)
	// Get the version of this Module
	// (GET /version)
	GetVersion(w http.ResponseWriter, r *http.Request)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// GetAuthorizationModelApiID operation middleware
func (siw *ServerInterfaceWrapper) GetAuthorizationModelApiID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAuthorizationModelApiID(w, r, apiID)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PutAuthorizationModelApiIDApprove operation middleware
func (siw *ServerInterfaceWrapper) PutAuthorizationModelApiIDApprove(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params PutAuthorizationModelApiIDApproveParams

	// ------------- Required query parameter "method" -------------
	if paramValue := r.URL.Query().Get("method"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "method"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "method", r.URL.Query(), &params.Method)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "method", Err: err})
		return
	}

	// ------------- Required query parameter "path" -------------
	if paramValue := r.URL.Query().Get("path"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "path"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "path", r.URL.Query(), &params.Path)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "path", Err: err})
		return
	}

	// ------------- Required query parameter "k8sClientUid" -------------
	if paramValue := r.URL.Query().Get("k8sClientUid"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "k8sClientUid"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "k8sClientUid", r.URL.Query(), &params.K8sClientUid)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "k8sClientUid", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PutAuthorizationModelApiIDApprove(w, r, apiID, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PutAuthorizationModelApiIDDeny operation middleware
func (siw *ServerInterfaceWrapper) PutAuthorizationModelApiIDDeny(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params PutAuthorizationModelApiIDDenyParams

	// ------------- Required query parameter "method" -------------
	if paramValue := r.URL.Query().Get("method"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "method"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "method", r.URL.Query(), &params.Method)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "method", Err: err})
		return
	}

	// ------------- Required query parameter "path" -------------
	if paramValue := r.URL.Query().Get("path"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "path"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "path", r.URL.Query(), &params.Path)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "path", Err: err})
		return
	}

	// ------------- Required query parameter "k8sClientUid" -------------
	if paramValue := r.URL.Query().Get("k8sClientUid"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "k8sClientUid"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "k8sClientUid", r.URL.Query(), &params.K8sClientUid)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "k8sClientUid", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PutAuthorizationModelApiIDDeny(w, r, apiID, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PutAuthorizationModelApiIDLearningReset operation middleware
func (siw *ServerInterfaceWrapper) PutAuthorizationModelApiIDLearningReset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params PutAuthorizationModelApiIDLearningResetParams

	// ------------- Required query parameter "nr_traces" -------------
	if paramValue := r.URL.Query().Get("nr_traces"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "nr_traces"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "nr_traces", r.URL.Query(), &params.NrTraces)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "nr_traces", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PutAuthorizationModelApiIDLearningReset(w, r, apiID, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PutAuthorizationModelApiIDLearningStart operation middleware
func (siw *ServerInterfaceWrapper) PutAuthorizationModelApiIDLearningStart(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params PutAuthorizationModelApiIDLearningStartParams

	// ------------- Required query parameter "nr_traces" -------------
	if paramValue := r.URL.Query().Get("nr_traces"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "nr_traces"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "nr_traces", r.URL.Query(), &params.NrTraces)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "nr_traces", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PutAuthorizationModelApiIDLearningStart(w, r, apiID, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PutAuthorizationModelApiIDLearningStop operation middleware
func (siw *ServerInterfaceWrapper) PutAuthorizationModelApiIDLearningStop(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PutAuthorizationModelApiIDLearningStop(w, r, apiID)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetEvent operation middleware
func (siw *ServerInterfaceWrapper) GetEvent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetEvent(w, r, id)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// PutEventIdOperation operation middleware
func (siw *ServerInterfaceWrapper) PutEventIdOperation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int

	err = runtime.BindStyledParameter("simple", false, "id", chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	// ------------- Path parameter "operation" -------------
	var operation OperationEnum

	err = runtime.BindStyledParameter("simple", false, "operation", chi.URLParam(r, "operation"), &operation)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "operation", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.PutEventIdOperation(w, r, id, operation)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetVersion operation middleware
func (siw *ServerInterfaceWrapper) GetVersion(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetVersion(w, r)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/authorizationModel/{apiID}", wrapper.GetAuthorizationModelApiID)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/authorizationModel/{apiID}/approve", wrapper.PutAuthorizationModelApiIDApprove)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/authorizationModel/{apiID}/deny", wrapper.PutAuthorizationModelApiIDDeny)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/authorizationModel/{apiID}/learning/reset", wrapper.PutAuthorizationModelApiIDLearningReset)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/authorizationModel/{apiID}/learning/start", wrapper.PutAuthorizationModelApiIDLearningStart)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/authorizationModel/{apiID}/learning/stop", wrapper.PutAuthorizationModelApiIDLearningStop)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/event/{id}", wrapper.GetEvent)
	})
	r.Group(func(r chi.Router) {
		r.Put(options.BaseURL+"/event/{id}/{operation}", wrapper.PutEventIdOperation)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/version", wrapper.GetVersion)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+xY3W7jNhN9FYLfd0nEafdm4TutrWbVXf/AstMCi8BgpHHMjUxqSSqta/jdC1KSRdm0",
	"42ySol3kKpY4PDNz5sxQ4QYnYpULDlwr3N1glSxhRe3PYByFD8B1wLnQVDPB7etcihykZmCfbhcZjTXV",
	"hX36v4QF7uL/dRrQToXY+fDL56Cy3BKcgtKMW9RP79Xo9isk+jGEneEEFiWGhkRDOlMgH9vbd223BMOf",
	"GiSnmdmn1zngLr4VIgPKzaoShUzgOwPbEizhW8EkpLj7xWXIcXuzJTjI2QRULrgC4yEFlUiWG0pwFwcc",
	"CYuJ9JJqxBSSoAvJEeOIZhlKqAKFxAItKMsKCeoCk73arEApegdOikpLxu8OQqwNb0htWLo2VASFXgrJ",
	"/rKlGogUskMRZEAlN8BeMo1lIx+mYfWoVg6djmoQg1j5oFLStS1XDsnUvjsNG9d2+wTsAFrRkiax85gJ",
	"ipQBT+CQIVrZGnc+joCn80KBPJ+ifUHvc3Ja4Pfv1Vw8X9tN2MRN0dX5WcQ11fUw13D6ndrZlcVD0wr0",
	"UqSeFiE4p3r5eO9Yqx0OaeL1Ze7MwO4GAy9WBmI4msfjsIcJ/hwGk2E0vLI/r6JpNAimISY4nsXjqBeN",
	"ZvF8EPaj2aD97mN09dHx16TQ35uQbWqZP2+Wz2maSlDKu1zORjf+X3+bYoI/BHFkcvg0Gl7Nf5/3RsN4",
	"Nggn86jvCW2PRGaYq4BbAZgx2VLgoT5ydg1SVdo5iPaecX+WnK7g6ILKaeJfLbyk7aVT2Hys68qRC0vc",
	"mH0q2XVDaAluiKZ5LsWDQUiBry2QfWF7sHpb/vaJIXamZCO9odHXeDK6jvphHxM8CU3lppNZbxr2vTgO",
	"3e1SPBytwx4/D0eTN5aML4TnPBxHvYxKptdoINIiAxSMI0ywZjqD9rrpMkyacPDlxeXFT9VRxGnOcBe/",
	"u7i8eIfLJrfBd+jB3OhsaM6i/tYs34GdlLvTIUpxF1+B9kwbs8ciS7oCbWf6lw1mJpBqWJTaw7SybKjR",
	"sgBSfYM5NDKu4c6M+e2NsS4/GWzUP19emj+J4Bq4Lvshz1hiw+l8VWU1GsCnDc6yHu06xEWSmM60Cwta",
	"ZPrlAnC+hzyeQymFRNKx2JJTZevU7WKEWnjKNy6OlS/YNdr5VTxRNVJt/FaAXDc7d2fGowJo+siPVEX0",
	"bJz796qXMeB6xp4W16vq8rQs/kOCtFP76Wrsl8P+TYpvUnwxKdb/2XQkqPJoe6IoP1cAE7v/ddXJ5VxL",
	"moD6Fx2WP54SlKbyOUqI7f43JfwIShD5s4Qg8pfSwVvldpWDB+C6s2HpyX9I7JXtWeyfd67+Q6Xw3DV7",
	"iAnGEbJ26A+ml4i2rAlWxWpF5bokAuklIGiMzaOzAVGeottFRpGqLqXbHHc2O2q3p3rBhhOlzQ3Wa1BP",
	"vCjC8Xkc7BTt7ZuGt25zus25zKharb2rVlhlh8QC6SVT1eWEe5Fcd2Z9b/KKHNcuPFleu3FCHaa/a47k",
	"ZJj5OwAA//+E60MNMBoAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
