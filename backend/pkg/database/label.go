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

package database

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	labelTableName    = "labels"
	evendIdColumnName = "event_id"
)

type Label struct {
	ID      uint   `gorm:"primarykey" faker:"-"`
	EventID uint   `json:"event_id" gorm:"column:event_id" faker:"-"`
	Key     string `json:"key" gorm:"column:key" faker:"-"`
	Value   string `json:"value" gorm:"column:value" faker:"-"`
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_label.go -package=database github.com/openclarity/apiclarity/backend/pkg/database LabelsTable
type LabelsTable interface {
	CreateLabels(ctx context.Context, eventID uint, labels map[string]string) error
	GetLabels(ctx context.Context, eventID uint) (map[string]string, error)
	DeleteLabels(ctx context.Context, eventID uint) error
}

func (l *LabelsTableHandler) CreateLabels(ctx context.Context, eventID uint, labels map[string]string) error {
	if len(labels) == 0 {
		return nil
	}

	//gorm can insert a slice so convert
	insertLabels := make([]Label, 0, len(labels))
	for k, v := range labels {
		insertLabels = append(insertLabels, Label{
			EventID: eventID,
			Key:     k,
			Value:   v,
		})
	}

	if result := l.tx.WithContext(ctx).Create(insertLabels); result.Error != nil {
		return fmt.Errorf("failed to create labels: %v", result.Error)
	} else {
		log.Debugf("labels created for event %d", eventID)
		return nil
	}
}

func (l *LabelsTableHandler) GetLabels(ctx context.Context, eventID uint) (map[string]string, error) {
	var labels []Label
	if err := l.tx.Where(Label{EventID: eventID}).WithContext(ctx).Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get labels for event %d: %v", eventID, err)
	}
	if len(labels) < 1 {
		return map[string]string{}, nil
	}

	labelMap := make(map[string]string, len(labels))
	for _, v := range labels {
		labelMap[v.Key] = v.Value
	}
	return labelMap, nil
}

func (l *LabelsTableHandler) DeleteLabels(ctx context.Context, eventID uint) error {
	return l.tx.Where(Label{EventID: eventID}).
		WithContext(ctx).
		Delete(&Label{}).
		Error
}

type LabelsTableHandler struct {
	tx *gorm.DB
}
