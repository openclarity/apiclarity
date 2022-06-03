package fuzzer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/core"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/clients"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/config"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/logging"
	"github.com/openclarity/apiclarity/backend/pkg/modules/internal/fuzzer/restapi"
)

const (
	ModuleName        = "fuzzer"
	ModuleDescription = "This is the Fuzzer module"
	ModuleVersion     = "0.0.0"
	EmptyJSON         = "{}"
	NbMaxServicePart  = 2
)

type pluginFuzzer struct {
	runningState restapi.TestingModuleState
	httpHandler  http.Handler

	config       *config.Config
	model        *Model
	fuzzerClient clients.Client

	accessor core.BackendAccessor
}

//nolint:gochecknoinits // was needed for the module implementation of ApiClarity
func init() {
	core.RegisterModule(newFuzzer)
}

func newFuzzer(ctx context.Context, accessor core.BackendAccessor) (core.Module, error) {
	logging.InitLogger()
	logging.Logf("[Fuzzer] Start():: -->")

	// Use default values
	plugin := pluginFuzzer{
		httpHandler:  nil,
		runningState: restapi.TestingModuleState{APIsInCache: 0, Version: ModuleVersion},
		config:       config.NewFuzzerConfig(),
		fuzzerClient: nil,
		model:        nil,
		accessor:     accessor,
	}

	plugin.config.Dump()

	plugin.httpHandler = restapi.HandlerWithOptions(&pluginFuzzerHTTPHandler{fuzzer: &plugin}, restapi.ChiServerOptions{BaseURL: core.BaseHTTPPath + "/" + ModuleName})

	// Initialize the model
	plugin.model = NewModel(accessor)
	err := plugin.model.Init(ctx)
	if err != nil {
		logging.Errorf("[Fuzzer] Error, failed to init model.")
	}
	logging.Logf("[Fuzzer] Model creation ok")

	// Create the client according to the configuration
	plugin.fuzzerClient, err = clients.NewClient(plugin.config, accessor)
	if err != nil {
		logging.Errorf("[testing] Error, failed to create a client")
		return nil, fmt.Errorf("ignoring fuzzer module due to missing fuzzer client")
	}
	logging.Logf("[testing] Client (%v) creation, ok", plugin.config.GetDeploymentType())

	logging.Logf("[Fuzzer] Start():: <--")

	return &plugin, nil
}

func (p *pluginFuzzer) Name() string {
	return ModuleName
}

func (p *pluginFuzzer) EventNotify(ctx context.Context, event *core.Event) {
	// Fuzzer doesn't use this
	// Logf("[Fuzzer] received a new trace for API(%s) EventID(%v)", event.APIInfoID, event.ID)
}

/*
*
*  Implement Fuzzer module stuff
*
 */

func (p *pluginFuzzer) FuzzTarget(ctx context.Context, apiID uint32, params restapi.FuzzTargetParams) error {
	// Check for deployment
	if p.fuzzerClient == nil {
		return &PluginError{"No deployment client running"}
	}

	// Retrieve the API (it will give the endpoint and the port)
	api, err := p.model.GetAPI(ctx, uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget():: can't retrieve API (%v)", apiID)
		return &NotFoundError{msg: ""}
	}

	logging.Logf("[Fuzzer] FuzzTarget():: API_id (%v) => API (%v) with parameters (%v)", apiID, api, DumpHTTPFuzzParam(params))

	// Construct the URI of the enpoint to fuzz
	serviceToTest := api.name
	if len(api.namespace) > 0 {
		serviceToTest = fmt.Sprintf("%s.%s", serviceToTest, api.namespace)
	} else if params.Service != nil {
		serviceToTest = *params.Service
		sp := strings.Split(serviceToTest, ".")
		if len(sp) > NbMaxServicePart {
			logging.Logf("[Fuzzer] FuzzTarget():: Service is bad formated (%v). Fuzz aborted!", params.Service)
			// Retur an n error
			return &InvalidParameterError{}
		}
	}
	sURI := fmt.Sprintf("http://%s:%v", serviceToTest, api.port)

	// Get auth material, if provided
	securityParam := ""
	if params.Type != nil && *params.Type != "NONE" {
		securityParam, err = GetAuthStringFromParam(params)
		if err != nil {
			logging.Errorf("[Fuzzer] FuzzTarget():: can't get auth material for (%v)", apiID)
			return &PluginError{msg: err.Error()}
		}
	}

	// Fuzz it!

	err = p.model.StartAPIFuzzing(uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget():: can't start fuzzing for API (%v)", apiID)
		return &PluginError{msg: err.Error()}
	}

	err = p.fuzzerClient.TriggerFuzzingJob(apiID, sURI, securityParam)
	if err != nil {
		logging.Errorf("[Fuzzer] FuzzTarget():: can't trigger fuzzing job for API (%v)", apiID)
		_ = p.model.StopAPIFuzzing(uint(apiID))
		return &PluginError{msg: err.Error()}
	}

	// Success
	return nil
}

type pluginFuzzerHTTPHandler struct {
	fuzzer *pluginFuzzer
}

func httpError(writer http.ResponseWriter, err error) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(writer).Encode(map[string]interface{}{"error": err.Error()}); err != nil {
		httpError(writer, err)
	}
}

func httpResponse(writer http.ResponseWriter, statusCode int, data interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	if err := json.NewEncoder(writer).Encode(data); err != nil {
		httpError(writer, err)
	}
}

//
// Return the version for the fuzzer module.
//
func (*pluginFuzzerHTTPHandler) GetVersion(writer http.ResponseWriter, req *http.Request) {
	logging.Logf("[Fuzzer] GetVersion(): -->")
	if err := json.NewEncoder(writer).Encode(restapi.Version{Version: ModuleVersion}); err != nil {
		httpError(writer, err)
	}
}

func (p *pluginFuzzerHTTPHandler) GetState(writer http.ResponseWriter, req *http.Request) {
	state := p.fuzzer.runningState
	httpResponse(writer, http.StatusOK, state)
}

//
// Launch a fuzzing for an API.
//
func (p *pluginFuzzerHTTPHandler) FuzzTarget(writer http.ResponseWriter, req *http.Request, apiID uint32, params restapi.FuzzTargetParams) {
	logging.Logf("[Fuzzer] FuzzTarget(%v, %v): -->", apiID, DumpHTTPFuzzParam(params))

	err := p.fuzzer.FuzzTarget(req.Context(), apiID, params)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")
		//nolint: errorlint // no wrapped error here
		switch err2 := err.(type) {
		case *NotFoundError:
			httpResponse(writer, http.StatusNotFound, EmptyJSON)
		case *InvalidParameterError:
			httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		case *PluginError:
			httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		case *NotSupportedError:
			httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		default:
			logging.Errorf("[Fuzzer] FuzzTarget(%v): unexpected error=(%v)", apiID, err2)
			httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		}
		return
	}

	// Success
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}

//
// Retrieve the last update status for the API.
//
func (p *pluginFuzzerHTTPHandler) GetUpdateStatus(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Logf("[Fuzzer] GetUpdateStatus(%v): -->", apiID)

	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] GetUpdateStatus(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}

	result := api.GetLastStatus()
	if result == nil {
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetUpdateStatus(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

//
// Receive last status update.
//
func (p *pluginFuzzerHTTPHandler) PostUpdateStatus(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Logf("[Fuzzer] PostUpdateStatus(%v): -->", apiID)

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): failed to read the request body, error=%v", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}

	logging.Logf("[Fuzzer] PostUpdateStatus(%v): Received a request of size=(%v)", apiID, len(body))
	/*
	* Decode the result
	 */
	var data restapi.FuzzingStatusAndReport
	err = json.Unmarshal(body, &data)
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): failed to decode the request body, error=%v", apiID, err)
		httpResponse(writer, http.StatusInternalServerError, EmptyJSON)
		return
	}
	// Logf("body=%v", data)

	/*
	* Add the new status to the last Test
	 */
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] PostUpdateStatus(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	// Logf("[Fuzzer] PostUpdateStatus():: API_id (%v) => API (%v)", apiId, api)
	api.AddNewStatusReport(data)
	// If the status indicate a completion, close the job
	if data.Progress == 100 && data.Status != "IN_PROGRESS" {
		err = p.fuzzer.model.StopAPIFuzzing(uint(apiID))
		if err != nil {
			logging.Errorf("[Fuzzer] PostUpdateStatus(%v): failed to stop fuzzing status, error=%v", apiID, err)
		}
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}

//
// Return the Raw findings with SCN format.
//
func (p *pluginFuzzerHTTPHandler) GetRawfindings(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Logf("[Fuzzer] GetRawfindings(%v): -->", apiID)
	httpResponse(writer, http.StatusNotImplemented, EmptyJSON)
}

//
// Return the findings list for the lastest Test.
//
func (p *pluginFuzzerHTTPHandler) GetFindings(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Logf("[Fuzzer] GetFindings(%v): -->", apiID)
	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		logging.Errorf("[Fuzzer] GetFindings(%v): Can't retrieve api_id=(%v), error=(%v)", apiID, apiID, err)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}
	logging.Logf("[Fuzzer] GetFindings(%v): API_id (%v) => API (%v)", apiID, apiID, api)
	var count int
	result := restapi.Findings{
		Items: api.GetLastFindings(),
		Total: &count,
	}
	count = len(*(result.Items))
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetFindings(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

//
// Receive findings for last Test.
//
func (p *pluginFuzzerHTTPHandler) PostRawfindings(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Logf("[Fuzzer] PostRawfindings(%v): -->", apiID)
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Errorf("[Fuzzer] PostRawfindings(%v): can't read body content, error=(%v)", apiID, err)
		httpResponse(writer, http.StatusBadRequest, EmptyJSON)
		return
	}
	log.Println(string(body))
	// p.fuzzer.accessor.StoreAPIInfoAnnotations(r.Context(), p.fuzzer.Name(), uint(apiID), core.Annotation{Name: "Fuzzer findings", Annotation: body})
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNoContent)
}

//
// Send the list of Tests for the API.
//
func (p *pluginFuzzerHTTPHandler) GetTests(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Debugf("[Fuzzer] GetTests(%v): -->", apiID)

	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		fmt.Printf("[Fuzzer] GetTests(%v):: can't retrieve API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	// Logf("[Fuzzer] GetTests():: API_id (%v) => API (%v)", apiID, api)

	count := 0
	// tests := api.tests
	result := restapi.Tests{
		Items: api.GetTestsList(),
		Total: &count,
	}
	count = len(*(result.Items))

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetTests(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

func (p *pluginFuzzerHTTPHandler) GetReport(writer http.ResponseWriter, req *http.Request, apiID uint32, timestamp int64) {
	logging.Logf("[Fuzzer] GetReport(%v): -->", apiID)

	api, err := p.fuzzer.model.GetAPI(req.Context(), uint(apiID))
	if err != nil {
		fmt.Printf("[Fuzzer] GetTests(%v): can't retrieve API (%v)", apiID, apiID)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	result := api.GetReport(timestamp)
	if result == nil {
		fmt.Printf("[Fuzzer] GetTests(%v): can't retrieve Report with timestamp (%v)", apiID, timestamp)
		httpResponse(writer, http.StatusNotFound, EmptyJSON)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	err = json.NewEncoder(writer).Encode(result)
	if err != nil {
		logging.Errorf("[Fuzzer] GetReport(%v): Failed to encode response, error=(%v)", apiID, err)
	}
}

func (p *pluginFuzzerHTTPHandler) GetAnnotatedSpec(writer http.ResponseWriter, req *http.Request, apiID uint32) {
	logging.Logf("[Fuzzer] GetAnnotatedSpec(): called for API_id (%v)", apiID)
	httpResponse(writer, http.StatusNotImplemented, EmptyJSON)
}

func (p *pluginFuzzer) HTTPHandler() http.Handler {
	return p.httpHandler
}
