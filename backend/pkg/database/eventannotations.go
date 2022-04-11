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

	"gorm.io/gorm"
)

const (
	eventAnnotationsTableName = "event_annotations"
	eventIDColumnName         = "event_id"
)

var alertKinds = []string{alertAnnotation}

type APIEventAnnotation struct {
	// will be populated after inserting to DB
	ID         uint   `gorm:"primarykey" faker:"-"`
	ModuleName string `json:"module_name,omitempty" gorm:"column:module_name;uniqueIndex:api_event_ann_idx_model" faker:"-"`
	EventID    uint   `json:"event_id,omitempty" gorm:"column:event_id;uniqueIndex:api_event_ann_idx_model" faker:"-"`
	Name       string `json:"name,omitempty" gorm:"column:name;uniqueIndex:api_event_ann_idx_model" faker:"-"`
	Annotation []byte `json:"annotation,omitempty" gorm:"column:annotation" faker:"-"`
}

type APIEventAnnotationTable interface {
	Create(ctx context.Context, eas ...APIEventAnnotation) error
	Get(ctx context.Context, modName string, eventID uint, name string) (*APIEventAnnotation, error)
	List(ctx context.Context, modName string, eventID uint) ([]*APIEventAnnotation, error)
}

type APIEventAnnotationTableHandler struct {
	tx *gorm.DB
}

func (APIEventAnnotation) TableName() string {
	return eventAnnotationsTableName
}

func (ea *APIEventAnnotationTableHandler) Create(ctx context.Context, annotations ...APIEventAnnotation) error {
	return ea.tx.WithContext(ctx).Create(annotations).Error
}

func (ea *APIEventAnnotationTableHandler) List(ctx context.Context, modName string, eventID uint) ([]*APIEventAnnotation, error) {
	var events []*APIEventAnnotation

	t := ea.tx.Where("module_name = ? AND event_id = ? AND name NOT IN ?", modName, eventID, alertKinds)

	if err := t.WithContext(ctx).Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (ea *APIEventAnnotationTableHandler) Get(ctx context.Context, modName string, eventID uint, name string) (*APIEventAnnotation, error) {
	annotation := &APIEventAnnotation{}

	t := ea.tx.Where("module_name = ? AND event_id = ? AND name = ?", modName, eventID, name)

	if err := t.WithContext(ctx).First(&annotation).Error; err != nil {
		return annotation, err
	}

	return annotation, nil
}
