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
	"strings"
	"time"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
	"github.com/openclarity/apiclarity/backend/pkg/utils"
	speculatorspec "github.com/openclarity/speculator/pkg/spec"
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

	Annotations []*APIEventAnnotation `gorm:"foreignKey:EventID;references:ID;constraint:OnDelete:CASCADE"`
	APIInfo     APIInfo               `gorm:"constraint:OnDelete:CASCADE"`
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_apievent.go -package=database github.com/openclarity/apiclarity/backend/pkg/database APIEventsTable
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
	UpdateAPIEvent(event *APIEvent) error
	GroupByAPIInfo() ([]HostGroup, error)
}

func (a *APIEventsTableHandler) UpdateAPIEvent(event *APIEvent) error {
	err := a.tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&event).Error

	return err
}

type GetAPIEventsQuery struct {
	EventID *uint32

	Offset     int
	Limit      int
	Order      string
	AscSortDir bool

	APIEventsFilters          *APIEventsFilters
	APIEventAnnotationFilters *APIEventAnnotationFilters
}

type APIEventAnnotationFilters struct {
	NoAnnotations bool

	ModuleNameIs *string
	NameIs       *string
	ValueIs      *string

	ModuleNameIsNot *string
	NameIsNot       *string
	ValueIsNot      *string
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
	APIInfoIDIs           *uint32
}

const dashboardTopAPIsNum = 5

func (a *APIEventsTableHandler) GetAPIEventsWithAnnotations(ctx context.Context, query GetAPIEventsQuery) ([]*APIEvent, error) {
	var events []*APIEvent
	if query.Order == "" {
		query.Order = timeColumnName
	}

	tx := a.tx
	if query.APIEventsFilters != nil {
		tx = a.setAPIEventsFilters(query.APIEventsFilters)
	}
	if query.EventID != nil {
		tx = tx.Where(fmt.Sprintf("%s.%s = ?", apiEventTableName, idColumnName), *query.EventID)
	}
	if query.APIEventAnnotationFilters != nil {
		var getInJSONMap func(key, value, extract string) string
		var args []interface{}

		switch tx.Dialector.(type) {
		case *postgres.Dialector:
			getInJSONMap = func(key, val, extract string) string {
				args = append(args, extract)
				return fmt.Sprintf("jsonb_object_agg(coalesce(%s, ''), %s) -> ?", key, val)
			}
		case *sqlite.Dialector:
			getInJSONMap = func(key, val, extract string) string {
				args = append(args, extract)
				return fmt.Sprintf("json_extract(json_group_object(coalesce(%s, ''), %s), ?)", key, val)
			}
		}
		var havingConditions []string
		if query.APIEventAnnotationFilters.NameIs != nil {
			havingConditions = append(havingConditions, getInJSONMap("ea."+nameColumnName, "true", *query.APIEventAnnotationFilters.NameIs)+" IS NOT NULL")
		}
		if query.APIEventAnnotationFilters.ModuleNameIs != nil {
			havingConditions = append(havingConditions, getInJSONMap("ea."+moduleNameColumnName, "true", *query.APIEventAnnotationFilters.ModuleNameIs)+" IS NOT NULL")
		}
		if query.APIEventAnnotationFilters.ValueIs != nil {
			havingConditions = append(havingConditions, getInJSONMap("ea."+annotationColumnName, "true", *query.APIEventAnnotationFilters.ValueIs)+" IS NOT NULL")
		}
		if query.APIEventAnnotationFilters.NameIsNot != nil {
			havingConditions = append(havingConditions, getInJSONMap("ea."+nameColumnName, "true", *query.APIEventAnnotationFilters.NameIsNot)+" IS NULL")
		}
		if query.APIEventAnnotationFilters.ModuleNameIsNot != nil {
			havingConditions = append(havingConditions, getInJSONMap("ea."+moduleNameColumnName, "true", *query.APIEventAnnotationFilters.ModuleNameIsNot)+" IS NULL")
		}
		if query.APIEventAnnotationFilters.ValueIsNot != nil {
			havingConditions = append(havingConditions, getInJSONMap("ea."+annotationColumnName, "true", *query.APIEventAnnotationFilters.ValueIsNot)+" IS NULL")
		}
		tx = tx.Joins(fmt.Sprintf("LEFT JOIN %s ea ON %s.%s = ea.%s ",
			eventAnnotationsTableName, apiEventTableName, idColumnName, eventIDColumnName)).
			Group(fmt.Sprintf("%s.%s", apiEventTableName, idColumnName)).
			Distinct()
		if query.APIEventAnnotationFilters.NoAnnotations {
			tx.Having(fmt.Sprintf("(%s) OR COUNT(ea.id) = 0",
				strings.Join(havingConditions, " AND ")), args...)
		} else {
			tx.Having(strings.Join(havingConditions, " AND "), args...)
		}
	}

	tx = tx.Preload("Annotations")

	sortDir := "DESC"
	if query.AscSortDir {
		sortDir = "ASC"
	}

	if query.Limit != 0 {
		tx = tx.Limit(query.Limit)
	}

	if err := tx.Offset(query.Offset).
		Order(fmt.Sprintf("%s %s", query.Order, sortDir)).
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
			Alert:      models.AlertSeverityEnum(ann.Name),
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
	// if the user requests a filter on alert we need to join with the event_annotations table
	// and do the filtering on the foreign table
	if len(params.AlertIs) > 0 || len(params.AlertTypeIs) > 0 {
		tx = tx.Joins(fmt.Sprintf("LEFT JOIN %s ea ON %s.%s = ea.%s",
			eventAnnotationsTableName,
			apiEventTableName, idColumnName,
			eventIDColumnName)).
			Distinct()
		if len(params.AlertTypeIs) > 0 {
			tx = tx.Where(fmt.Sprintf("ea.%s IN ?", moduleNameColumnName), params.AlertTypeIs).Where(fmt.Sprintf("ea.%s = 'ALERT'", nameColumnName))
		}
		if len(params.AlertIs) > 0 {
			or := tx.Where(fmt.Sprintf("ea.%s LIKE ?", annotationColumnName), params.AlertIs[0])
			for severity := range params.AlertIs[1:] {
				or = or.Or(fmt.Sprintf("ea.%s LIKE ?", annotationColumnName), severity)
			}
			tx = tx.Where(or)
		}
	}
	// get total count item with the current filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
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

	// API Id filter
	tx = FilterIsUint32(tx, apiInfoIDColumnName, filters.APIInfoIDIs)

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
		APIInfoIDIs:          params.APIInfoIDIs,
	}
}

func FilterIsNotInAnnotations(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("(SELECT COUNT(%s) FROM %s WHERE %s = %s.%s AND %s IN ?) = 0",
		idColumnName,
		eventAnnotationsTableName,
		eventIDColumnName, apiEventTableName, idColumnName, column), values)
}
