// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/utils"
	speculatorspec "github.com/apiclarity/speculator/pkg/spec"
)

const (
	apiEventTableName = "api_events"

	// NOTE: when changing one of the column names change also the gorm label in APIEvent.
	timeColumnName                 = "time"
	requestTimeColumnName          = "request_time"
	methodColumnName               = "method"
	pathColumnName                 = "path"
	providedPathIDColumnName       = "provided_path_id"
	reconstructedPathIDColumnName  = "reconstructed_path_id"
	statusCodeColumnName           = "status_code"
	sourceIPColumnName             = "source_ip"
	destinationIPColumnName        = "destination_ip"
	destinationPortColumnName      = "destination_port"
	hasSpecDiffColumnName          = "has_spec_diff" // hasProvidedSpecDiff || hasReconstructedSpecDiff
	specDiffTypeColumnName         = "spec_diff_type"
	hostSpecNameColumnName         = "host_spec_name"
	newReconstructedSpecColumnName = "new_reconstructed_spec"
	oldReconstructedSpecColumnName = "old_reconstructed_spec"
	newProvidedSpecColumnName      = "new_provided_spec"
	oldProvidedSpecColumnName      = "old_provided_spec"
	apiInfoIDColumnName            = "api_info_id"
	isNonAPIColumnName             = "is_non_api"
	eventTypeColumnName            = "event_type"
)

const alertAnnotation = "ALERT"

var specDiffColumns = []string{newReconstructedSpecColumnName, oldReconstructedSpecColumnName, newProvidedSpecColumnName, oldProvidedSpecColumnName}

type APIEvent struct {
	// will be populated after inserting to DB
	ID uint `gorm:"primarykey" faker:"-"`
	// CreatedAt time.Time
	// UpdatedAt time.Time

	Time                     strfmt.DateTime   `json:"time" gorm:"column:time" faker:"-"`
	RequestTime              strfmt.DateTime   `json:"requestTime" gorm:"column:request_time" faker:"-"`
	Method                   models.HTTPMethod `json:"method,omitempty" gorm:"column:method" faker:"oneof: GET, PUT, POST, DELETE"`
	Path                     string            `json:"path,omitempty" gorm:"column:path" faker:"oneof: /news, /customers, /jokes"`
	ProvidedPathID           string            `json:"providedPathId,omitempty" gorm:"column:provided_path_id" faker:"-"`
	ReconstructedPathID      string            `json:"reconstructedPathId,omitempty" gorm:"column:reconstructed_path_id" faker:"-"`
	Query                    string            `json:"query,omitempty" gorm:"column:query" faker:"oneof: name=ferret&color=purple, foo=bar, -"`
	StatusCode               int64             `json:"statusCode,omitempty" gorm:"column:status_code" faker:"oneof: 200, 401, 404, 500"`
	SourceIP                 string            `json:"sourceIP,omitempty" gorm:"column:source_ip" faker:"sourceIP"`
	DestinationIP            string            `json:"destinationIP,omitempty" gorm:"column:destination_ip" faker:"destinationIP"`
	DestinationPort          int64             `json:"destinationPort,omitempty" gorm:"column:destination_port" faker:"oneof: 80, 443"`
	HasReconstructedSpecDiff bool              `json:"hasReconstructedSpecDiff,omitempty" gorm:"column:has_reconstructed_spec_diff"`
	HasProvidedSpecDiff      bool              `json:"hasProvidedSpecDiff,omitempty" gorm:"column:has_provided_spec_diff"`
	HasSpecDiff              bool              `json:"hasSpecDiff,omitempty" gorm:"column:has_spec_diff"`
	SpecDiffType             models.DiffType   `json:"specDiffType,omitempty" gorm:"column:spec_diff_type" faker:"oneof: ZOMBIE_DIFF, SHADOW_DIFF, GENERAL_DIFF, NO_DIFF"`
	HostSpecName             string            `json:"hostSpecName,omitempty" gorm:"column:host_spec_name" faker:"oneof: test.com, example.com, kaki.org"`
	IsNonAPI                 bool              `json:"isNonApi,omitempty" gorm:"column:is_non_api" faker:"-"`

	// Spec diff info
	// New reconstructed spec json string
	NewReconstructedSpec string `json:"newReconstructedSpec,omitempty" gorm:"column:new_reconstructed_spec" faker:"-"`
	// Old reconstructed spec json string
	OldReconstructedSpec string `json:"oldReconstructedSpec,omitempty" gorm:"column:old_reconstructed_spec" faker:"-"`
	// New provided spec json string
	NewProvidedSpec string `json:"newProvidedSpec,omitempty" gorm:"column:new_provided_spec" faker:"-"`
	// Old provided spec json string
	OldProvidedSpec string `json:"oldProvidedSpec,omitempty" gorm:"column:old_provided_spec" faker:"-"`

	// ID for the relevant APIInfo
	APIInfoID uint `json:"apiInfoId,omitempty" gorm:"column:api_info_id" faker:"-"`
	// We'll not always have a corresponding API info (e.g. non-API resources) so the type is needed also for the event
	EventType models.APIType `json:"eventType,omitempty" gorm:"column:event_type" faker:"oneof: INTERNAL, EXTERNAL"`

	Annotations []*APIEventAnnotation `gorm:"foreignKey:EventID;references:ID"`
}

type APIEventsTable interface {
	GetAPIEventsWithAnnotations(ctx context.Context, filters GetAPIEventsQuery) ([]*APIEvent, error)
	GetAPIEventsAndTotal(params operations.GetAPIEventsParams) ([]APIEvent, int64, error)
	GetAPIEvent(eventID uint32) (*APIEvent, error)
	GetAPIEventReconstructedSpecDiff(eventID uint32) (*APIEvent, error)
	GetAPIEventProvidedSpecDiff(eventID uint32) (*APIEvent, error)
	SetAPIEventsReconstructedPathID(approvedReview []*speculatorspec.ApprovedSpecReviewPathItem, host string, port string) error
	GetAPIEventsLatestDiffs(latestDiffsNum int) ([]APIEvent, error)
	GetAPIUsages(params operations.GetAPIUsageHitCountParams) ([]*models.HitCount, error)
	GetDashboardAPIUsages(startTime, endTime time.Time, apiType APIUsageType) ([]*models.APIUsage, error)
	CreateAPIEvent(event *APIEvent)
	GroupByAPIInfo() ([]HostGroup, error)
}

type GetAPIEventsQuery struct {
	EventID *uint32

	Offset int
	Limit  int
	Order  string

	Filters           *APIEventsFilters
	AnnotationFilters *AnnotationFilters
}

type AnnotationFilters struct {
	ModuleNameIs []string
	NameIs       []string
	ValueIs      []string

	ModuleNameIsNot []string
	NameIsNot       []string
	ValueIsNot      []string
}

type APIEventAnnotationFilter struct {
	Names       []string
	ModuleNames []string
	Annotations []string
}

type APIEventsTableHandler struct {
	tx *gorm.DB
}

type HostGroup struct {
	HostSpecName string
	Port         int64
	APIType      string
	APIInfoID    uint32
	Count        int
}

type APIEventsFilters struct {
	DestinationIPIsNot    []string
	DestinationIPIs       []string
	DestinationPortIsNot  []string
	DestinationPortIs     []string
	EndTime               *strfmt.DateTime
	RequestEndTime        *strfmt.DateTime
	ShowNonAPI            bool
	HasSpecDiffIs         *bool
	SpecDiffTypeIs        []string
	MethodIs              []string
	ReconstructedPathIDIs []string
	ProvidedPathIDIs      []string
	PathContains          []string
	PathEnd               *string
	PathIsNot             []string
	PathIs                []string
	PathStart             *string
	SourceIPIsNot         []string
	SourceIPIs            []string
	SpecContains          []string
	SpecEnd               *string
	SpecIsNot             []string
	SpecIs                []string
	SpecStart             *string
	StartTime             *strfmt.DateTime
	RequestStartTime      *strfmt.DateTime
	StatusCodeGte         *string
	StatusCodeIsNot       []string
	StatusCodeIs          []string
	StatusCodeLte         *string
}

const dashboardTopAPIsNum = 5

func (a *APIEventsTableHandler) GetAPIEventsWithAnnotations(ctx context.Context, query GetAPIEventsQuery) ([]*APIEvent, error) {
	var events []*APIEvent
	if query.Order == "" {
		query.Order = timeColumnName
	}

	tx := a.tx
	if query.Filters != nil {
		tx = a.setAPIEventsFilters(query.Filters)
	}
	if query.EventID != nil {
		tx = tx.Where(fmt.Sprintf("%s.%s = ?", apiEventTableName, idColumnName), *query.EventID)
	}
	if query.AnnotationFilters != nil {
		tx = FilterIs(tx, "ea."+nameColumnName, query.AnnotationFilters.NameIs)
		tx = FilterIs(tx, "ea."+moduleNameColumnName, query.AnnotationFilters.ModuleNameIs)
		tx = FilterIs(tx, "ea."+annotationColumnName, query.AnnotationFilters.ValueIs)

		tx = FilterIsNotOrNull(tx, "ea."+nameColumnName, query.AnnotationFilters.NameIsNot)
		tx = FilterIsNotOrNull(tx, "ea."+moduleNameColumnName, query.AnnotationFilters.ModuleNameIsNot)
		tx = FilterIsNotOrNull(tx, "ea."+annotationColumnName, query.AnnotationFilters.ValueIsNot)
	}
	tx = tx.Joins(fmt.Sprintf("LEFT JOIN %s ea ON %s.%s = ea.%s ",
		eventAnnotationsTableName, apiEventTableName, idColumnName, eventIDColumnName)).
		Distinct()

	tx = tx.Preload("Annotations")
	if err := tx.Offset(query.Offset).
		Limit(query.Limit).
		Order(fmt.Sprintf("%s DESC", query.Order)).
		WithContext(ctx).
		Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (a *APIEventsTableHandler) GroupByAPIInfo() ([]HostGroup, error) {
	var results []HostGroup

	rows, err := a.tx.
		// filters out non APIs
		Not(isNonAPIColumnName+" = ?", true).
		Select(
			FieldInTable(apiEventTableName, hostSpecNameColumnName) +
				", " + FieldInTable(apiEventTableName, destinationPortColumnName) +
				", " + FieldInTable(apiEventTableName, apiInfoIDColumnName) +
				", " + FieldInTable(apiEventTableName, eventTypeColumnName) +
				", COUNT(*) AS count").
		Group(FieldInTable(apiEventTableName, hostSpecNameColumnName)).
		Group(FieldInTable(apiEventTableName, destinationPortColumnName)).
		Group(FieldInTable(apiEventTableName, apiInfoIDColumnName)).
		Group(FieldInTable(apiEventTableName, eventTypeColumnName)).
		Order("count desc").
		Limit(dashboardTopAPIsNum).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get top API event counts: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Warnf("Failed to close rows: %v", err)
		}
	}()

	for rows.Next() {
		group := HostGroup{}
		if err := rows.Scan(&group.HostSpecName, &group.Port, &group.APIInfoID, &group.APIType, &group.Count); err != nil {
			return nil, fmt.Errorf("failed to get fields: %v", err)
		}
		log.Debugf("Fetched fields: %+v", group)
		results = append(results, group)
	}

	return results, nil
}

func (APIEvent) TableName() string {
	return apiEventTableName
}

func APIEventFromDB(event *APIEvent) *models.APIEvent {
	e := &models.APIEvent{
		APIInfoID:                uint32(event.APIInfoID),
		APIType:                  event.EventType,
		DestinationIP:            event.DestinationIP,
		DestinationPort:          event.DestinationPort,
		HasProvidedSpecDiff:      &event.HasProvidedSpecDiff,
		HasReconstructedSpecDiff: &event.HasReconstructedSpecDiff,
		HostSpecName:             event.HostSpecName,
		ID:                       uint32(event.ID),
		Method:                   event.Method,
		Path:                     event.Path,
		Query:                    event.Query,
		SourceIP:                 event.SourceIP,
		SpecDiffType:             &event.SpecDiffType,
		StatusCode:               event.StatusCode,
		Time:                     event.Time,
		RequestTime:              event.RequestTime,
		Alerts:                   []*models.ModuleAlert{},
	}
	for _, ann := range event.Annotations {
		e.Alerts = append(e.Alerts, &models.ModuleAlert{
			Alert:      ann.Name,
			ModuleName: ann.ModuleName,
			Reason:     string(ann.Annotation),
		})
	}
	return e
}

func (a *APIEventsTableHandler) CreateAPIEvent(event *APIEvent) {
	if result := a.tx.Create(event); result.Error != nil {
		log.Errorf("Failed to create event: %v", result.Error)
	} else {
		log.Debugf("Event created %+v", event)
	}
}

func (a *APIEventsTableHandler) GetAPIEventsAndTotal(params operations.GetAPIEventsParams) ([]APIEvent, int64, error) {
	var apiEvents []APIEvent
	var count int64

	tx := a.setAPIEventsFilters(getAPIEventsParamsToFilters(params))

	// get total count item with the current filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if len(params.AlertIs) > 0 {
		tx = tx.Joins(fmt.Sprintf("LEFT JOIN %s ea ON %s.%s = ea.%s",
			eventAnnotationsTableName,
			apiEventTableName, idColumnName,
			eventIDColumnName)).
			Where(fmt.Sprintf("ea.%s IN ?", annotationColumnName), params.AlertIs).
			Distinct()
	}
	sortOrder, err := CreateSortOrder(params.SortKey, params.SortDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create sort order: %v", err)
	}

	tx = tx.Scopes(Paginate(params.Page, params.PageSize)).
		Preload("Annotations", fmt.Sprintf("%s = ?", nameColumnName), alertAnnotation)

	// get specific page ordered items with the current filters
	if err := tx.Order(sortOrder).
		Omit(specDiffColumns...).
		Find(&apiEvents).Error; err != nil {
		return nil, 0, err
	}

	return apiEvents, count, nil
}

func (a *APIEventsTableHandler) GetAPIEvent(eventID uint32) (*APIEvent, error) {
	var apiEvent APIEvent

	tx := a.tx
	tx = tx.Preload("Annotations", fmt.Sprintf("%s = ?", nameColumnName), alertAnnotation)
	if err := tx.Omit(specDiffColumns...).First(&apiEvent, eventID).Error; err != nil {
		return nil, err
	}

	return &apiEvent, nil
}

func (a *APIEventsTableHandler) GetAPIEventReconstructedSpecDiff(eventID uint32) (*APIEvent, error) {
	var apiEvent APIEvent

	if err := a.tx.Select(newReconstructedSpecColumnName, oldReconstructedSpecColumnName, specDiffTypeColumnName).First(&apiEvent, eventID).Error; err != nil {
		return nil, err
	}

	return &apiEvent, nil
}

func (a *APIEventsTableHandler) GetAPIEventProvidedSpecDiff(eventID uint32) (*APIEvent, error) {
	var apiEvent APIEvent

	if err := a.tx.Select(newProvidedSpecColumnName, oldProvidedSpecColumnName, specDiffTypeColumnName).First(&apiEvent, eventID).Error; err != nil {
		return nil, err
	}

	return &apiEvent, nil
}

func (a *APIEventsTableHandler) GetAPIEventsLatestDiffs(latestDiffsNum int) ([]APIEvent, error) {
	var latestDiffs []APIEvent
	if err := a.tx.Where(hasSpecDiffColumnName + " = true").
		Order("time desc").Limit(latestDiffsNum).Scan(&latestDiffs).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest diffs from events table. %v", err)
	}

	return latestDiffs, nil
}

func (a *APIEventsTableHandler) setAPIEventsFilters(filters *APIEventsFilters) *gorm.DB {
	tx := a.tx
	if filters.StartTime != nil && filters.EndTime != nil {
		tx = tx.Where(CreateTimeFilter(timeColumnName, *filters.StartTime, *filters.EndTime))
	}
	if filters.RequestStartTime != nil && filters.RequestEndTime != nil {
		tx = tx.Where(CreateTimeFilter(requestTimeColumnName, *filters.RequestStartTime, *filters.RequestEndTime))
	}

	// methods filter
	tx = FilterIs(tx, methodColumnName, filters.MethodIs)

	// path ID filters
	tx = FilterIs(tx, providedPathIDColumnName, filters.ProvidedPathIDIs)
	tx = FilterIs(tx, reconstructedPathIDColumnName, filters.ReconstructedPathIDIs)

	// path filters
	tx = FilterIs(tx, pathColumnName, filters.PathIs)
	tx = FilterIsNot(tx, pathColumnName, filters.PathIsNot)
	tx = FilterContains(tx, pathColumnName, filters.PathContains)
	tx = FilterStartsWith(tx, pathColumnName, filters.PathStart)
	tx = FilterEndsWith(tx, pathColumnName, filters.PathEnd)

	// status codes filters
	tx = FilterIs(tx, statusCodeColumnName, filters.StatusCodeIs)
	tx = FilterIsNot(tx, statusCodeColumnName, filters.StatusCodeIsNot)
	tx = FilterGte(tx, statusCodeColumnName, filters.StatusCodeGte)
	tx = FilterLte(tx, statusCodeColumnName, filters.StatusCodeLte)

	// source IPs filters
	tx = FilterIs(tx, sourceIPColumnName, filters.SourceIPIs)
	tx = FilterIsNot(tx, sourceIPColumnName, filters.SourceIPIsNot)
	// destination IPs filters
	tx = FilterIs(tx, destinationIPColumnName, filters.DestinationIPIs)
	tx = FilterIsNot(tx, destinationIPColumnName, filters.DestinationIPIsNot)
	// destination ports filters
	tx = FilterIs(tx, destinationPortColumnName, filters.DestinationPortIs)
	tx = FilterIsNot(tx, destinationPortColumnName, filters.DestinationPortIsNot)

	// has spec diff filter
	tx = FilterIsBool(tx, hasSpecDiffColumnName, filters.HasSpecDiffIs)

	// spec diff type filter
	tx = FilterIs(tx, specDiffTypeColumnName, filters.SpecDiffTypeIs)

	// host spec name filters
	tx = FilterIs(tx, hostSpecNameColumnName, filters.SpecIs)
	tx = FilterIsNot(tx, hostSpecNameColumnName, filters.SpecIsNot)
	tx = FilterContains(tx, hostSpecNameColumnName, filters.SpecContains)
	tx = FilterStartsWith(tx, hostSpecNameColumnName, filters.SpecStart)
	tx = FilterEndsWith(tx, hostSpecNameColumnName, filters.SpecEnd)

	// ignore non APIs
	if !filters.ShowNonAPI {
		tx.Where(fmt.Sprintf("%s = ?", isNonAPIColumnName), false)
	}

	return tx
}

// SetAPIEventsReconstructedPathID will set reconstructed path ID for all events with the provided paths, host and port.
func (a *APIEventsTableHandler) SetAPIEventsReconstructedPathID(approvedReview []*speculatorspec.ApprovedSpecReviewPathItem, host string, port string) error {
	err := a.tx.Transaction(func(tx *gorm.DB) error {
		for _, item := range approvedReview {
			tx := FilterIs(tx, pathColumnName, utils.MapToSlice(item.Paths))
			tx = FilterIs(tx, hostSpecNameColumnName, []string{host})
			tx = FilterIs(tx, destinationPortColumnName, []string{port})

			if err := tx.Model(&APIEvent{}).Updates(map[string]interface{}{reconstructedPathIDColumnName: item.PathUUID}).Error; err != nil {
				// return any error will rollback
				return err
			}
		}

		// return nil will commit the whole transaction
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to set API events path ID: %v", err)
	}

	return nil
}

func getAPIUsageHitCountParamsToFilters(params operations.GetAPIUsageHitCountParams) *APIEventsFilters {
	return &APIEventsFilters{
		DestinationIPIsNot:    params.DestinationIPIsNot,
		DestinationIPIs:       params.DestinationIPIs,
		DestinationPortIsNot:  params.DestinationPortIsNot,
		DestinationPortIs:     params.DestinationPortIs,
		EndTime:               &params.EndTime,
		ShowNonAPI:            params.ShowNonAPI,
		HasSpecDiffIs:         params.HasSpecDiffIs,
		SpecDiffTypeIs:        params.SpecDiffTypeIs,
		MethodIs:              params.MethodIs,
		ReconstructedPathIDIs: params.ReconstructedPathIDIs,
		ProvidedPathIDIs:      params.ProvidedPathIDIs,
		PathContains:          params.PathContains,
		PathEnd:               params.PathEnd,
		PathIsNot:             params.PathIsNot,
		PathIs:                params.PathIs,
		PathStart:             params.PathStart,
		SourceIPIsNot:         params.SourceIPIsNot,
		SourceIPIs:            params.SourceIPIs,
		SpecContains:          params.SpecContains,
		SpecEnd:               params.SpecEnd,
		SpecIsNot:             params.SpecIsNot,
		SpecIs:                params.SpecIs,
		SpecStart:             params.SpecStart,
		StartTime:             &params.StartTime,
		StatusCodeGte:         params.StatusCodeGte,
		StatusCodeIsNot:       params.StatusCodeIsNot,
		StatusCodeIs:          params.StatusCodeIs,
		StatusCodeLte:         params.StatusCodeLte,
	}
}

func getAPIEventsParamsToFilters(params operations.GetAPIEventsParams) *APIEventsFilters {
	return &APIEventsFilters{
		DestinationIPIsNot:   params.DestinationIPIsNot,
		DestinationIPIs:      params.DestinationIPIs,
		DestinationPortIsNot: params.DestinationPortIsNot,
		DestinationPortIs:    params.DestinationPortIs,
		EndTime:              &params.EndTime,
		ShowNonAPI:           params.ShowNonAPI,
		HasSpecDiffIs:        params.HasSpecDiffIs,
		SpecDiffTypeIs:       params.SpecDiffTypeIs,
		MethodIs:             params.MethodIs,
		PathContains:         params.PathContains,
		PathEnd:              params.PathEnd,
		PathIsNot:            params.PathIsNot,
		PathIs:               params.PathIs,
		PathStart:            params.PathStart,
		SourceIPIsNot:        params.SourceIPIsNot,
		SourceIPIs:           params.SourceIPIs,
		SpecContains:         params.SpecContains,
		SpecEnd:              params.SpecEnd,
		SpecIsNot:            params.SpecIsNot,
		SpecIs:               params.SpecIs,
		SpecStart:            params.SpecStart,
		StartTime:            &params.StartTime,
		StatusCodeGte:        params.StatusCodeGte,
		StatusCodeIsNot:      params.StatusCodeIsNot,
		StatusCodeIs:         params.StatusCodeIs,
		StatusCodeLte:        params.StatusCodeLte,
	}
}
