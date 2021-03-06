// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/client-go/kubernetes"

	"github.com/openclarity/apiclarity/backend/pkg/database"
	pluginsmodels "github.com/openclarity/apiclarity/plugins/api/server/models"
)

type Annotation struct {
	Name       string
	Annotation []byte
}

type Event struct {
	APIEvent  *database.APIEvent
	Telemetry *pluginsmodels.Telemetry
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_modules.go -package=core github.com/openclarity/apiclarity/backend/pkg/modules/internal/core Module,BackendAccessor

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

func NewAccessor(dbHandler *database.Handler, clientset kubernetes.Interface) BackendAccessor {
	return &accessor{dbHandler, clientset}
}

type accessor struct {
	dbHandler *database.Handler
	clientset kubernetes.Interface
}

func (b *accessor) K8SClient() kubernetes.Interface {
	return b.clientset
}

func (b *accessor) GetAPIInfo(ctx context.Context, apiID uint) (*database.APIInfo, error) {
	apiInfo := &database.APIInfo{}
	if err := b.dbHandler.APIInventoryTable().First(apiInfo, apiID); err != nil {
		return nil, fmt.Errorf("failed to retrieve API info for apiID=%v: %v", apiID, err)
	}
	return apiInfo, nil
}

func (b *accessor) GetAPIEvents(ctx context.Context, filter database.GetAPIEventsQuery) ([]*database.APIEvent, error) {
	events, err := b.dbHandler.APIEventsTable().GetAPIEventsWithAnnotations(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("unable to get apievents with annotations: %w", err)
	}
	return events, nil
}

func (b *accessor) GetAPIEventAnnotation(ctx context.Context, modName string, eventID uint, name string) (*Annotation, error) {
	ann, err := b.dbHandler.APIEventsAnnotationsTable().Get(ctx, modName, eventID, name)
	if err != nil {
		return nil, fmt.Errorf("unable to get apievent annotation: %w", err)
	}
	return &Annotation{Name: ann.Name, Annotation: ann.Annotation}, nil
}

func (b *accessor) ListAPIEventAnnotations(ctx context.Context, modName string, eventID uint) ([]*Annotation, error) {
	var anns []*Annotation
	dbAnnotations, err := b.dbHandler.APIEventsAnnotationsTable().List(ctx, modName, eventID)
	if err != nil {
		return nil, fmt.Errorf("unable to list apievent annotations: %w", err)
	}
	for _, ann := range dbAnnotations {
		anns = append(anns, &Annotation{Name: ann.Name, Annotation: ann.Annotation})
	}
	return anns, nil
}

func (b *accessor) CreateAPIEventAnnotations(ctx context.Context, modName string, eventID uint, annotations ...Annotation) error {
	var dbAnns []database.APIEventAnnotation

	for _, a := range annotations {
		dbAnns = append(dbAnns, database.APIEventAnnotation{
			EventID:    eventID,
			ModuleName: modName,
			Name:       a.Name,
			Annotation: a.Annotation,
		})
	}

	if err := b.dbHandler.APIEventsAnnotationsTable().Create(ctx, dbAnns...); err != nil {
		return fmt.Errorf("unable to create apiinfo annotation: %w", err)
	}
	return nil
}

func (b *accessor) GetAPIInfoAnnotation(ctx context.Context, modName string, apiID uint, name string) (*Annotation, error) {
	ann, err := b.dbHandler.APIInfoAnnotationsTable().Get(ctx, modName, apiID, name)
	if err != nil {
		return nil, fmt.Errorf("unable to get apiinfo annotation: %w", err)
	}
	return &Annotation{Name: ann.Name, Annotation: ann.Annotation}, nil
}

func (b *accessor) ListAPIInfoAnnotations(ctx context.Context, modName string, apiID uint) (annotations []*Annotation, err error) {
	anns, err := b.dbHandler.APIInfoAnnotationsTable().List(ctx, modName, apiID)
	if err != nil {
		return nil, fmt.Errorf("unable to list apiinfo annotation: %w", err)
	}
	for _, ann := range anns {
		annotations = append(annotations, &Annotation{
			Name:       ann.Name,
			Annotation: ann.Annotation,
		})
	}
	return annotations, nil
}

func (b *accessor) StoreAPIInfoAnnotations(ctx context.Context, modName string, apiID uint, annotations ...Annotation) error {
	var dbAnns []database.APIInfoAnnotation

	for _, a := range annotations {
		dbAnns = append(dbAnns, database.APIInfoAnnotation{
			APIID:      apiID,
			ModuleName: modName,
			Name:       a.Name,
			Annotation: a.Annotation,
		})
	}

	if err := b.dbHandler.APIInfoAnnotationsTable().UpdateOrCreate(ctx, dbAnns...); err != nil {
		return fmt.Errorf("unable to store apiinfo annotation: %w", err)
	}
	return nil
}

func (b *accessor) DeleteAPIInfoAnnotations(ctx context.Context, modName string, apiID uint, name ...string) error {
	if err := b.dbHandler.APIInfoAnnotationsTable().Delete(ctx, modName, apiID, name...); err != nil {
		return fmt.Errorf("unable to delete the apiinfo annotation: %w", err)
	}
	return nil
}
