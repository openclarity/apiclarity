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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	apiEventAnnotationsTableName = "api_annotations"
	apiIdColumnName              = "api_id"
	moduleNameColumnName         = "module_name"
	annotationColumnName         = "annotation"
)

type APIInfoAnnotation struct {
	ID         uint   `gorm:"primarykey" faker:"-"`
	ModuleName string `json:"module_name,omitempty" gorm:"column:module_name;uniqueIndex:api_ann_idx_model" faker:"-"`
	APIID      uint   `json:"api_id,omitempty" gorm:"column:api_id;uniqueIndex:api_ann_idx_model" faker:"-"`
	Name       string `json:"name,omitempty" gorm:"column:name;uniqueIndex:api_ann_idx_model" faker:"-"`

	Annotation []byte `json:"annotation,omitempty" gorm:"column:annotation" faker:"-"`
}

type APIAnnotationsTable interface {
	UpdateOrCreate(ctx context.Context, am ...APIInfoAnnotation) error
	Get(ctx context.Context, modName string, apiID uint, name string) (*APIInfoAnnotation, error)
	List(ctx context.Context, modName string, apiID uint) ([]*APIInfoAnnotation, error)
	Delete(ctx context.Context, modName string, apiID uint, names ...string) error
}

type APIInfoAnnotationsTableHandler struct {
	tx *gorm.DB
}

func (APIInfoAnnotation) TableName() string {
	return apiEventAnnotationsTableName
}

func (am *APIInfoAnnotationsTableHandler) UpdateOrCreate(ctx context.Context, annotations ...APIInfoAnnotation) error {
	return am.tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: moduleNameColumnName}, {Name: apiIdColumnName}, {Name: nameColumnName}},
		UpdateAll: true,
	}).WithContext(ctx).Create(&annotations).Error
}

func (am *APIInfoAnnotationsTableHandler) Get(ctx context.Context, modName string, apiID uint, name string) (*APIInfoAnnotation, error) {
	var model APIInfoAnnotation

	if err := am.tx.Where(fmt.Sprintf("%s = ? AND %s = ? AND %s = ?",
		moduleNameColumnName, apiIdColumnName, nameColumnName), modName, apiID, name).
		WithContext(ctx).
		First(&model).
		Error; err != nil {
		return nil, err
	}

	return &model, nil
}

func (am *APIInfoAnnotationsTableHandler) List(ctx context.Context, modName string, apiID uint) ([]*APIInfoAnnotation, error) {
	var annotations []*APIInfoAnnotation

	var t *gorm.DB
	if modName == "" {
		t = am.tx.Where(fmt.Sprintf("%s = ?", apiIdColumnName), apiID)
	} else {
		t = am.tx.Where(fmt.Sprintf("%s = ? AND %s = ?", moduleNameColumnName, apiIdColumnName), modName, apiID)
	}
	if err := t.WithContext(ctx).Find(&annotations).Error; err != nil {
		return nil, err
	}

	return annotations, nil
}

func (am *APIInfoAnnotationsTableHandler) Delete(ctx context.Context, modName string, apiID uint, names ...string) error {
	return am.tx.Where(fmt.Sprintf("%s = ? AND %s = ? AND %s IN ?",
		moduleNameColumnName, apiIdColumnName, nameColumnName), modName, apiID, names).
		WithContext(ctx).
		Delete(&APIInfoAnnotation{}).
		Error
}
