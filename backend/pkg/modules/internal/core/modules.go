package core

import (
	"context"
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/apiclarity/apiclarity/backend/pkg/database"
	pluginsmodels "github.com/apiclarity/apiclarity/plugins/api/server/models"
)

type Annotation struct {
	Name       string
	Annotation []byte
}

type Event struct {
	APIEvent  *database.APIEvent
	Telemetry *pluginsmodels.Telemetry
}

// Module each APIClarity module needs to implement this interface.
type Module interface {
	Name() string

	// EventNotify called when a new API Request/reply is received by APIClarity.
	EventNotify(ctx context.Context, event *Event)

	// HTTPHandler that will be served by APIClarity under /api/modules/{moduleName}
	HTTPHandler() http.Handler
}

type BackendAccessor interface {
	K8SClient() kubernetes.Interface

	GetAPIInfo(ctx context.Context, apiID uint) (*database.APIInfo, error)
	GetAPIEvents(ctx context.Context, filter database.GetAPIEventsQuery) ([]*database.APIEvent, error)

	GetAPIEventAnnotation(ctx context.Context, modName string, eventID uint, name string) (*Annotation, error)
	ListAPIEventAnnotations(ctx context.Context, modName string, eventID uint) ([]*Annotation, error)
	CreateAPIEventAnnotations(ctx context.Context, modName string, eventID uint, annotations ...Annotation) error

	GetAPIInfoAnnotation(ctx context.Context, modName string, apiID uint, name string) (*Annotation, error)
	ListAPIInfoAnnotations(ctx context.Context, modName string, apiID uint) ([]*Annotation, error)
	StoreAPIInfoAnnotations(ctx context.Context, modName string, apiID uint, annotations ...Annotation) error
	DeleteAPIInfoAnnotations(ctx context.Context, modName string, apiID uint, name ...string) error
}
