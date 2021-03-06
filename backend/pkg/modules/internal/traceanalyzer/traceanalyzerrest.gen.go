// Package traceanalyzer provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.1 DO NOT EDIT.
package traceanalyzer

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

// Annotation defines model for Annotation.
type Annotation struct {
	Annotation string `json:"annotation"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Severity   string `json:"severity"`
}

// Annotations defines model for Annotations.
type Annotations struct {
	Items *[]Annotation `json:"items,omitempty"`

	// Total event annotations count
	Total int `json:"total"`
}

// DeleteAPIAnnotationsParams defines parameters for DeleteAPIAnnotations.
type DeleteAPIAnnotationsParams struct {
	// name of the annotation
	Name string `json:"name"`
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Delete Annotations for an API
	// (DELETE /apiAnnotations/{apiID})
	DeleteAPIAnnotations(w http.ResponseWriter, r *http.Request, apiID int64, params DeleteAPIAnnotationsParams)
	// Get Annotations for an API
	// (GET /apiAnnotations/{apiID})
	GetAPIAnnotations(w http.ResponseWriter, r *http.Request, apiID int64)
	// Get Annotations for an event
	// (GET /eventAnnotations/{eventID})
	GetEventAnnotations(w http.ResponseWriter, r *http.Request, eventID int64)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// DeleteAPIAnnotations operation middleware
func (siw *ServerInterfaceWrapper) DeleteAPIAnnotations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int64

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params DeleteAPIAnnotationsParams

	// ------------- Required query parameter "name" -------------
	if paramValue := r.URL.Query().Get("name"); paramValue != "" {

	} else {
		siw.ErrorHandlerFunc(w, r, &RequiredParamError{ParamName: "name"})
		return
	}

	err = runtime.BindQueryParameter("form", true, true, "name", r.URL.Query(), &params.Name)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "name", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeleteAPIAnnotations(w, r, apiID, params)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetAPIAnnotations operation middleware
func (siw *ServerInterfaceWrapper) GetAPIAnnotations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "apiID" -------------
	var apiID int64

	err = runtime.BindStyledParameter("simple", false, "apiID", chi.URLParam(r, "apiID"), &apiID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "apiID", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetAPIAnnotations(w, r, apiID)
	}

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler(w, r.WithContext(ctx))
}

// GetEventAnnotations operation middleware
func (siw *ServerInterfaceWrapper) GetEventAnnotations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "eventID" -------------
	var eventID int64

	err = runtime.BindStyledParameter("simple", false, "eventID", chi.URLParam(r, "eventID"), &eventID)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "eventID", Err: err})
		return
	}

	var handler = func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.GetEventAnnotations(w, r, eventID)
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
		r.Delete(options.BaseURL+"/apiAnnotations/{apiID}", wrapper.DeleteAPIAnnotations)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/apiAnnotations/{apiID}", wrapper.GetAPIAnnotations)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/eventAnnotations/{eventID}", wrapper.GetEventAnnotations)
	})

	return r
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+RUMW/bPBD9K8R93yhYbhN00KY2ReDNQLMFGVj6ZDOVSOZ4MqAa/O8FKcWWLAXI0KJD",
	"J0s6+r13793xBMo2zho07KE4gVcHbGR6LI2xLFlbE98cWYfEGlNNTmrcOYQCPJM2ewgZ/NBmt1gwssHF",
	"gscjkuZuoRgyIHxpNeEOikfQOxhwsrGMEcTA/5S9Itnvz6g40lx68vOmNGMzffifsIIC/ssvJuWDQ/nI",
	"nnBmkkSyS++WZR0hdugVaddbBQ/xs8AjGhYX8V4o2xqGM4w2jHukWe896ryxeE6bys4Jy+3mSy2jLWJb",
	"t3ttSlFuN5FJc41LBz5DBkck3/9/vVqvPsSGrEMjnYYCblbr1Q1k4CQfkku5dHpkbH6STm/uQi+mRsa5",
	"rHtk8UBSoSiNrLufSGKEICpLQgrvUOlKq0FxjCrVNzso4C4hl9vNONIoimSDjOSheDyBjmRR6OvMFJDE",
	"wdhXphazYfKj1MpSI7mP4dPtYirX/URsYSvBBxSToUwCXlqk7qJgGN63BVwvwFM87J01vh/Uj+vbuaXf",
	"WqXQ+6qtRXI9DWbcrLZpJHVny+ZGm2RwyGCP/Ceiukf+SznNjVvHH2UNo0m9SudqrZKs/Nn319mF4X37",
	"7/sFvFq80fUwSSH6+VYEIYM8XQ6TdUpfhoX6DREluKWQvl4xvyumQdy/E1RvXwgh/AoAAP//NX1lKzwH",
	"AAA=",
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
