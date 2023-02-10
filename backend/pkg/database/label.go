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

	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	servermodels "github.com/openclarity/apiclarity/api/server/models"
	models "github.com/openclarity/apiclarity/api3/common"
	apilabels "github.com/openclarity/apiclarity/plugins/api/labels"
)

const (
	labelTableName      = "labels"
	evendIdColumnName   = "event_id"
	apiInfoIdColumnName = "api_info_id"
)

type Label struct {
	ID        uint `gorm:"primaryKey" faker:"-"`
	APIInfoID uint `json:"api_info_id" gorm:"column:api_info_id,index" faker:"-"`
	//Since the APIInfoID does not identify an operation, we need to put that here
	Path    string            `json:"path,omitempty" gorm:"column:path" faker:"oneof: /news, /customers, /jokes"`
	Method  models.HttpMethod `json:"method,omitempty" gorm:"column:method" faker:"oneof: GET, PUT, POST, DELETE"`
	EventID uint              `json:"event_id" gorm:"column:event_id,index" faker:"-"`
	Key     string            `json:"key" gorm:"column:key" faker:"-"`
	Value   string            `json:"value" gorm:"column:value" faker:"-"`
}

//go:generate $GOPATH/bin/mockgen -destination=./mock_label.go -package=database github.com/openclarity/apiclarity/backend/pkg/database LabelsTable
type LabelsTable interface {
	CreateLabels(ctx context.Context, event *APIEvent, labels map[string]string) error
	GetLabelsByEventID(ctx context.Context, eventID uint) (map[string]string, error)
	DeleteLabels(ctx context.Context, eventID uint) error
	GetLabelsLineageChildren(ctx context.Context, apiID uint, method *models.HttpMethod, path *string) ([]Label, error)
	GetLabelsLineageParents(ctx context.Context, apiID uint, method *models.HttpMethod, path *string) ([]Label, error)
	ReplaceLabelMatching(ctx context.Context, key string, currentValue string, newValue string) (int64, error)
}

// Yep, this is nonsense due generating packages from multiple specs.
func convertHTTPMethodToHttpMethod(method servermodels.HTTPMethod) (models.HttpMethod, error) {
	if err := method.Validate(strfmt.Default); err != nil {
		return models.GET, err
	}
	switch method {
	case servermodels.HTTPMethodGET:
		return models.GET, nil
	case servermodels.HTTPMethodCONNECT:
		return models.CONNECT, nil
	case servermodels.HTTPMethodDELETE:
		return models.DELETE, nil
	case servermodels.HTTPMethodHEAD:
		return models.HEAD, nil
	case servermodels.HTTPMethodOPTIONS:
		return models.OPTIONS, nil
	case servermodels.HTTPMethodPATCH:
		return models.PATCH, nil
	case servermodels.HTTPMethodPOST:
		return models.POST, nil
	case servermodels.HTTPMethodPUT:
		return models.PUT, nil
	case servermodels.HTTPMethodTRACE:
		return models.TRACE, nil
	default:
		return models.GET, fmt.Errorf("unknown servermodel http method: %v", method)
	}
}

func (l *LabelsTableHandler) CreateLabels(ctx context.Context, event *APIEvent, labels map[string]string) error {
	if len(labels) == 0 {
		return nil
	}
	eventID := event.ID
	path := event.Path
	if event.ProvidedPathID != "" {
		path = event.ProvidedPathID
	} else if event.ReconstructedPathID != "" {
		path = event.ReconstructedPathID
	}

	method, err := convertHTTPMethodToHttpMethod(event.Method)
	if err != nil {
		return fmt.Errorf("failed to convert http method: %v", err)
	}

	//gorm can insert a slice so convert
	insertLabels := make([]Label, 0, len(labels))
	for k, v := range labels {
		insertLabels = append(insertLabels, Label{
			EventID:   eventID,
			APIInfoID: event.APIInfoID,
			Path:      path,
			Method:    method,
			Key:       k,
			Value:     v,
		})
	}

	if result := l.tx.WithContext(ctx).Create(insertLabels); result.Error != nil {
		return fmt.Errorf("failed to create labels: %v", result.Error)
	} else {
		log.Debugf("labels created for event %d", eventID)
		return nil
	}
}

func (l *LabelsTableHandler) GetLabelsByEventID(ctx context.Context, eventID uint) (map[string]string, error) {
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

func (l *LabelsTableHandler) ReplaceLabelMatching(ctx context.Context, key string, currentValue string, newValue string) (int64, error) {
	result := l.tx.Where(Label{Key: key, Value: currentValue}).
		WithContext(ctx).
		Update("value", newValue)
	return result.RowsAffected, result.Error
}

func (l *LabelsTableHandler) GetLabelsLineageChildren(ctx context.Context, apiID uint, method *models.HttpMethod, path *string) ([]Label, error) {
	//Find children through parent relationship
	searchLabel := Label{APIInfoID: apiID, Key: apilabels.DataLineageIDKey}
	if method != nil {
		searchLabel.Method = *method
	}
	if (path != nil) && *path != "" {
		searchLabel.Path = *path
	}

	var labels []Label
	err := l.db.Where("Value = (?) AND Key = ?",
		l.tx.Select("Value").Where(searchLabel), //lineageID
		apilabels.DataLineageParentKey,
	).Distinct("APIInfoID", "Path", "Method").WithContext(ctx).Find(&labels).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get children lineage labels for apiID %d, with method %s, and path %s: %v", apiID, method, path, err)
	}
	return labels, nil
}

func (l *LabelsTableHandler) GetLabelsLineageParents(ctx context.Context, apiID uint, method *models.HttpMethod, path *string) ([]Label, error) {
	//Find parent through labels
	searchLabel := Label{APIInfoID: apiID, Key: apilabels.DataLineageParentKey}
	if method != nil {
		searchLabel.Method = *method
	}
	if (path != nil) && *path != "" {
		searchLabel.Path = *path
	}

	var labels []Label
	err := l.db.Where("Value = (?) AND Key = ?",
		l.tx.Select("Value").Where(searchLabel), //parentLineageID
		apilabels.DataLineageIDKey,
	).Distinct("APIInfoID", "Path", "Method").WithContext(ctx).Find(&labels).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get parent lineage labels for apiID %d, with method %s, and path %s: %v", apiID, method, path, err)
	} else {
		return labels, nil
	}
}

type LabelsTableHandler struct {
	tx *gorm.DB
	db *gorm.DB
}
