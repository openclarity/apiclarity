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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	apiAnnotationsTableName = "api_annotations"
)

type APIAnnotation struct {
	// will be populated after inserting to DB
	ID uint `gorm:"primarykey" faker:"-"`
	// CreatedAt time.Time
	// UpdatedAt time.Time

	ModuleName string `json:"module_name,omitempty" gorm:"column:module_name;uniqueIndex:api_ann_idx_model" faker:"-"`
	APIID      uint   `json:"api_id,omitempty" gorm:"column:api_id;uniqueIndex:api_ann_idx_model" faker:"-"`
	API        APIInfo

	Name       string `json:"name,omitempty" gorm:"column:name;uniqueIndex:api_ann_idx_model" faker:"-"`
	Annotation []byte `json:"annotation,omitempty" gorm:"column:annotation" faker:"-"`
}

type APIAnnotationsTable interface {
	UpdateOrCreate(am *APIAnnotation) error
	BulkCreate(aas *[]APIAnnotation) error
	GetAnnotation(modName string, apiID uint, name string) (*APIAnnotation, error)
	GetAnnotations(modName string, apiID uint) ([]APIAnnotation, error)
	DeleteAnnotations(modName string, apiID uint, annIDs []uint) error
}

type APIAnnotationsTableHandler struct {
	tx *gorm.DB
}

func (APIAnnotation) TableName() string {
	return apiAnnotationsTableName
}

func (am *APIAnnotationsTableHandler) UpdateOrCreate(annotation *APIAnnotation) error {
	am.tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "module_name"}, {Name: "api_id"}, {Name: "name"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"annotation": annotation.Annotation}),
	}).Create(&annotation)

	return nil
}

func (am *APIAnnotationsTableHandler) BulkCreate(annotations *[]APIAnnotation) error {
	am.tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "module_name"}, {Name: "api_id"}, {Name: "name"}},
		UpdateAll: true,
	}).Create(annotations)

	return nil
}

func (am *APIAnnotationsTableHandler) GetAnnotation(modName string, apiID uint, name string) (*APIAnnotation, error) {
	var model APIAnnotation

	t := am.tx.Where("module_name = ? AND api_id = ? AND name = ?", modName, apiID, name)
	if err := t.First(&model).Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (am *APIAnnotationsTableHandler) GetAnnotations(modName string, apiID uint) ([]APIAnnotation, error) {
	var annotations []APIAnnotation

	var t *gorm.DB
	if modName == "" {
		t = am.tx.Where("api_id = ?", apiID)
	} else {
		t = am.tx.Where("module_name = ? AND api_id = ?", modName, apiID)
	}
	if err := t.Find(&annotations).Error; err != nil {
		return nil, err
	}

	return annotations, nil
}

func (am *APIAnnotationsTableHandler) DeleteAnnotations(modName string, apiID uint, annIDs []uint) error {
	return am.tx.Where("api_id = ?", apiID).Delete(&APIAnnotation{}, annIDs).Error
}
