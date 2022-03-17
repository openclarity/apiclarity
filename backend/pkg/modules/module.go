package module

import (
	"net/http"

	"github.com/apiclarity/apiclarity/api/server/models"
	_database "github.com/apiclarity/apiclarity/backend/pkg/database"
	pluginsmodels "github.com/apiclarity/apiclarity/plugins/api/server/models"
)

type Annotation struct {
	Id         uint32
	Name       string
	Annotation []byte
}

type AlertSeverity int

const BaseHTTPPath = "/api/modules"

const (
	NONE AlertSeverity = iota
	HIGH
	CRITICAL
)
const (
	ALERT_NONE     = "ALERT_NONE"
	ALERT_HIGH     = "ALERT_HIGH"
	ALERT_CRITICAL = "ALERT_CRITICAL"
)

func (es AlertSeverity) String() string {
	switch es {
	case NONE:
		return ALERT_NONE
	case HIGH:
		return ALERT_HIGH
	case CRITICAL:
		return ALERT_CRITICAL
	}
	return ALERT_NONE
}

type APIEvent = _database.APIEvent
type APIInfo = _database.APIInfo
type GetEventAnnotationFilter = _database.GetEventAnnotationFilter
type EventAnnotation = _database.EventAnnotation

// The order of the modules is not important.
// You MUST NOT rely on a specific order of modules.
var modules = &[]Module{}

func Modules() *[]Module {
	return modules
}

func AddModule(m Module) {
	*modules = append(*modules, m)
}

// Module each APIClarity module needs to implement this interface.
type Module interface {
	Name() string
	Description() string

	Start(ctx BackendModuleWrapper) error
	Stop() error

	// EventNotify called when a new API Request/reply is received by APIClarity.
	EventNotify(event APIEvent, trace *pluginsmodels.Telemetry) error

	// EventAnnotationNotify called when a module set a new Event Annotation
	EventAnnotationNotify(modName string, eventID uint, ann Annotation) error

	// APIAnnotationNotify called when an API model is published by a module for a given API.
	// For example, when the API Trace Analyzer module publishes a CRUD model, any other
	// interested plugin is notified.
	APIAnnotationNotify(modName string, apiID uint, annotation *Annotation) error

	// HTTPHandler that will be served by APIClarity under /api/modules/{moduleName}
	HTTPHandler() http.Handler
}

type BackendModuleWrapper interface {
	SetEventAnnotation(modName string, eventID uint, annotation Annotation) error
	SetEventAnnotations(modName string, eventID uint, annotations []Annotation) error
	GetEventAnnotation(modName string, eventID uint, name string) (Annotation, error)
	GetEventAnnotations(modName string, eventID uint) ([]Annotation, error)
	GetEventAnnotationsHistory(modName string, filter GetEventAnnotationFilter) ([]EventAnnotation, error)
	SetEventAlert(modName string, eventID uint, severity AlertSeverity) error

	SetAPIAnnotation(modName string, apiID uint, annotation Annotation) error
	SetAPIAnnotations(modName string, apiID uint, annotations []Annotation) error
	GetAPIAnnotation(modName string, apiID uint, name string) (Annotation, error)
	GetAPIAnnotations(modName string, apiID uint) ([]Annotation, error)
	DeleteAPIAnnotations(modName string, apiID uint, annIDs []uint) error

	GetAPISpecsInfo(apiID uint32) (*models.OpenAPISpecs, error)
	GetAPISpecs(apiID uint32) (*APIInfo, error)
	GetAPIEvent(eventID uint32) (*APIEvent, error)
	// This allow to get full API information from the backend to the modules. Just differ slighly that GetAPISpecs()
	GetAPI(apiID uint32) (*APIInfo, error)
}
