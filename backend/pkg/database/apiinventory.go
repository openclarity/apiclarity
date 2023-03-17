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
	"fmt"

	"github.com/go-openapi/strfmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api/server/restapi/operations"
)

const (
	apiInventoryTableName = "api_inventory"

	// NOTE: when changing one of the column names change also the gorm label in APIInfo.
	idColumnName                         = "id"
	typeColumnName                       = "type"
	nameColumnName                       = "name"
	portColumnName                       = "port"
	hasProvidedSpecColumnName            = "has_provided_spec"
	hasReconstructedSpecColumnName       = "has_reconstructed_spec"
	reconstructedSpecColumnName          = "reconstructed_spec"
	reconstructedSpecInfoColumnName      = "reconstructed_spec_info"
	providedSpecColumnName               = "provided_spec"
	providedSpecInfoColumnName           = "provided_spec_info"
	providedSpecCreatedAtColumnName      = "provided_spec_created_at"
	reconstructedSpecCreatedAtColumnName = "reconstructed_spec_created_at"
)

type APIInfo struct {
	// will be populated after inserting to DB
	ID uint `json:"id,omitempty" gorm:"primarykey" faker:"-"`

	Type                       models.APIType  `json:"type,omitempty" gorm:"column:type;uniqueIndex:api_info_idx_model" faker:"oneof: INTERNAL, EXTERNAL"`
	Name                       string          `json:"name,omitempty" gorm:"column:name;uniqueIndex:api_info_idx_model" faker:"oneof: test.com, example.com, kaki.org"`
	Port                       int64           `json:"port,omitempty" gorm:"column:port;uniqueIndex:api_info_idx_model" faker:"oneof: 80, 443"`
	TraceSourceID              uint            `json:"traceSourceID,omitempty" gorm:"column:trace_source_id;default:0;uniqueIndex:api_info_idx_model" faker:"-"` // This is the name of the Trace Source which notified of this API. Empty means it was auto discovered by APIClarity on first.
	HasProvidedSpec            bool            `json:"hasProvidedSpec,omitempty" gorm:"column:has_provided_spec"`
	HasReconstructedSpec       bool            `json:"hasReconstructedSpec,omitempty" gorm:"column:has_reconstructed_spec"`
	ReconstructedSpec          string          `json:"reconstructedSpec,omitempty" gorm:"column:reconstructed_spec" faker:"-"`
	ReconstructedSpecInfo      string          `json:"reconstructedSpecInfo,omitempty" gorm:"column:reconstructed_spec_info" faker:"-"`
	ProvidedSpec               string          `json:"providedSpec,omitempty" gorm:"column:provided_spec" faker:"-"`
	ProvidedSpecInfo           string          `json:"providedSpecInfo,omitempty" gorm:"column:provided_spec_info" faker:"-"`
	DestinationNamespace       string          `json:"destinationNamespace,omitempty" gorm:"column:destination_namespace;uniqueIndex:api_info_idx_model" faker:"-"`
	ProvidedSpecCreatedAt      strfmt.DateTime `json:"providedSpecCreatedAt,omitempty" gorm:"column:provided_spec_created_at" faker:"-"`
	ReconstructedSpecCreatedAt strfmt.DateTime `json:"reconstructedSpecCreatedAt,omitempty" gorm:"column:reconstructed_spec_created_at" faker:"-"`

	TraceSource TraceSource          `gorm:"constraint:OnDelete:CASCADE"`
	Annotations []*APIInfoAnnotation `gorm:"foreignKey:APIID;references:ID;constraint:OnDelete:CASCADE"`
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_apiinventory.go -package=database github.com/openclarity/apiclarity/backend/pkg/database APIInventoryTable
type APIInventoryTable interface {
	GetAPIInventoryAndTotal(params operations.GetAPIInventoryParams) ([]APIInfo, int64, error)
	GetAPISpecs(apiID uint32) (*APIInfo, error)
	GetAPISpecsInfo(apiID uint32) (*models.OpenAPISpecs, error)
	PutAPISpec(apiID uint, spec string, specInfo *models.SpecInfo, specType specType, createdAt strfmt.DateTime) error
	DeleteProvidedAPISpec(apiID uint32) error
	DeleteApprovedAPISpec(apiID uint32) error
	GetAPIID(name, port string, traceSourceID *uuid.UUID) (uint, error)
	First(dest *APIInfo, conds ...interface{}) error
	FirstOrCreate(apiInfo *APIInfo) (created bool, err error)
	CreateAPIInfo(event *APIInfo)
}

type APIInventoryTableHandler struct {
	tx *gorm.DB
}

func (APIInfo) TableName() string {
	return apiInventoryTableName
}

func APIInfoFromDB(apiInfo *APIInfo) *models.APIInfo {
	return &models.APIInfo{
		HasProvidedSpec:      &apiInfo.HasProvidedSpec,
		HasReconstructedSpec: &apiInfo.HasReconstructedSpec,
		ID:                   uint32(apiInfo.ID),
		Name:                 apiInfo.Name,
		Port:                 apiInfo.Port,
		DestinationNamespace: apiInfo.DestinationNamespace,
		TraceSourceID:        strfmt.UUID(apiInfo.TraceSource.UID.String()),
		TraceSourceName:      apiInfo.TraceSource.Name,
		TraceSourceType:      apiInfo.TraceSource.Type,
	}
}

func (a *APIInventoryTableHandler) CreateAPIInfo(event *APIInfo) {
	if result := a.tx.Create(event); result.Error != nil {
		log.Errorf("Failed to create api: %v", result.Error)
	} else {
		log.Infof("API created %+v", event)
	}
}

func (a *APIInventoryTableHandler) GetAPIInventoryAndTotal(params operations.GetAPIInventoryParams) ([]APIInfo, int64, error) {
	var apiInventory []APIInfo
	var count int64

	tx := a.setAPIInventoryFilters(params)
	// get total count item with the current filters
	if err := tx.Count(&count).Error; err != nil {
		return nil, 0, err
	}

	sortOrder, err := CreateSortOrder(params.SortKey, params.SortDir)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to create sort order: %v", err)
	}

	// get specific page ordered items with the current filters
	if err := tx.Scopes(Paginate(params.Page, params.PageSize)).
		Order(sortOrder).
		Preload("TraceSource").
		Find(&apiInventory).Error; err != nil {
		return nil, 0, err
	}

	return apiInventory, count, nil
}

func (a *APIInventoryTableHandler) setAPIInventoryFilters(params operations.GetAPIInventoryParams) *gorm.DB {
	table := a.tx
	// type filter
	table = FilterIs(table, typeColumnName, []string{params.Type})

	// id filter
	if params.APIID != nil {
		table = FilterIs(table, idColumnName, []string{*params.APIID})
	}

	// names filter
	table = FilterIs(table, nameColumnName, params.NameIs)
	table = FilterIsNot(table, nameColumnName, params.NameIsNot)
	table = FilterContains(table, nameColumnName, params.NameContains)
	table = FilterStartsWith(table, nameColumnName, params.NameStart)
	table = FilterEndsWith(table, nameColumnName, params.NameEnd)

	// ports filters
	table = FilterIs(table, portColumnName, params.PortIs)
	table = FilterIsNot(table, portColumnName, params.PortIsNot)

	// has provided spec diff filter
	table = FilterIsBool(table, hasProvidedSpecColumnName, params.HasProvidedSpecIs)

	// has reconstructed spec diff filter
	table = FilterIsBool(table, hasReconstructedSpecColumnName, params.HasReconstructedSpecIs)

	return table
}

func (a *APIInventoryTableHandler) GetAPIID(name, port string, traceSourceID *uuid.UUID) (uint, error) {
	apiInfo := APIInfo{}
	cond := map[string]interface{}{
		apiInventoryTableName + "." + nameColumnName: name,
		apiInventoryTableName + "." + portColumnName: port,
	}
	tx := a.tx.Where(cond)
	if traceSourceID != nil {
		tx.Joins("TraceSource").Where("uid = ?", *traceSourceID)
	}
	if result := tx.First(&apiInfo); result.Error != nil {
		return 0, result.Error
	}

	return apiInfo.ID, nil
}

func (a *APIInventoryTableHandler) First(dest *APIInfo, conds ...interface{}) error {
	return a.tx.Preload("TraceSource").First(dest, conds).Error
}

func (a *APIInventoryTableHandler) FirstOrCreate(apiInfo *APIInfo) (created bool, err error) {
	tx := a.tx.Preload("TraceSource").Where(*apiInfo).FirstOrCreate(apiInfo)
	return tx.RowsAffected > 0, tx.Error
}
