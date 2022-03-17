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
	"time"

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	eventAnnotationsTableName = "event_annotations"
)

var alertKinds = []string{"ALERT"}

type GetEventAnnotationFilter struct {
	StartTime        *time.Time
	EndTime          *time.Time
	StartRequestTime *time.Time
	EndRequestTime   *time.Time
	Name             *string
	Method           *string
	PathId           *string
	SpecType         *string
}

type EventAnnotation struct {
	// will be populated after inserting to DB
	ID uint `gorm:"primarykey" faker:"-"`
	// CreatedAt time.Time
	// UpdatedAt time.Time

	ModuleName string `json:"module_name,omitempty" gorm:"column:module_name" faker:"-"`
	EventID    uint   `json:"event_id,omitempty" gorm:"column:event_id" faker:"-"`
	Event      APIEvent

	Name       string `json:"name,omitempty" gorm:"column:name" faker:"-"`
	Annotation []byte `json:"annotation,omitempty" gorm:"column:annotation" faker:"-"`
}

type EventAnnotationsTable interface {
	Create(ea *EventAnnotation) error
	BulkCreate(eas *[]EventAnnotation) error
	GetAnnotations(modName string, eventID uint) ([]EventAnnotation, error)
	GetAnnotation(modName string, eventID uint, name string) (EventAnnotation, error)
	GetAnnotationsHistory(modName string, filter GetEventAnnotationFilter) ([]EventAnnotation, error)
	GetEventsAlerts(eventIDs []uint) ([]EventAnnotation, error)
}

type EventAnnotationsTableHandler struct {
	tx *gorm.DB
}

func (EventAnnotation) TableName() string {
	return eventAnnotationsTableName
}

func (ea *EventAnnotationsTableHandler) Create(annotation *EventAnnotation) error {
	if err := ea.tx.Create(annotation).Error; err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (ea *EventAnnotationsTableHandler) BulkCreate(annotations *[]EventAnnotation) error {
	if err := ea.tx.Create(annotations).Error; err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (ea *EventAnnotationsTableHandler) GetAnnotations(modName string, eventID uint) ([]EventAnnotation, error) {
	var events []EventAnnotation

	t := ea.tx.Where("module_name = ? AND event_id = ? AND name NOT IN ?", modName, eventID, alertKinds)

	if err := t.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (ea *EventAnnotationsTableHandler) GetAnnotation(modName string, eventID uint, name string) (EventAnnotation, error) {
	var annotation EventAnnotation

	t := ea.tx.Where("module_name = ? AND event_id = ? AND name = ?", modName, eventID, name)

	if err := t.First(&annotation).Error; err != nil {
		return annotation, err
	}

	return annotation, nil
}

func (ea *EventAnnotationsTableHandler) GetAnnotationsHistory(modName string, filter GetEventAnnotationFilter) ([]EventAnnotation, error) {
	var evAnnotations []EventAnnotation

	t := ea.tx.
		Joins("Event").
		Where("event_annotations.module_name = ?", modName)

	if filter.StartRequestTime != nil && filter.EndRequestTime != nil {
		t = t.Where(fmt.Sprintf("Event.request_time BETWEEN '%v' AND '%v'", strfmt.DateTime(*filter.StartRequestTime), strfmt.DateTime(*filter.EndRequestTime)))
	}
	if filter.StartTime != nil && filter.EndTime != nil {
		t = t.Where(fmt.Sprintf("Event.time BETWEEN '%v' AND '%v'", strfmt.DateTime(*filter.StartTime), strfmt.DateTime(*filter.EndTime)))
	}
	if filter.Method != nil {
		t = t.Where("Event.method = ?", *filter.Method)
	}
	if filter.Name != nil {
		t = t.Where("event_annotations.name = ?", *filter.Name)
	}
	if filter.PathId != nil && filter.SpecType != nil {
		pathIdField := ""
		if *filter.SpecType == "provided" {
			pathIdField = "Event.provided_path_id"
		} else {
			pathIdField = "Event.reconstructed_path_id"
		}
		t = t.Where(pathIdField+" = ?", *filter.PathId)
	}

	if err := t.Find(&evAnnotations).Error; err != nil {
		return nil, err
	}
	return evAnnotations, nil
}

func (ea *EventAnnotationsTableHandler) GetEventsAlerts(eventIDs []uint) ([]EventAnnotation, error) {
	var result []EventAnnotation

	t := ea.tx.Where("event_id IN ? AND name IN ?", eventIDs, alertKinds)

	if err := t.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}
