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
)

const (
	traceSamplingTableName = "trace_sampling"
	externalTraceSourceID  = 0
)

type TraceSampling struct {
	gorm.Model

	APIID         uint   `json:"api_id,omitempty" gorm:"column:api_id" faker:"-"`
	TraceSourceID uint   `json:"trace_source_id,omitempty" gorm:"column:trace_source_id" faker:"-"`
	Component     string `json:"component,omitempty" gorm:"column:component" faker:"-"`
}

type TraceSamplingTable interface {
	AddHostToTrace(apiID uint32, traceSourceID uint, component string) error
	GetHostsToTrace(traceSourceID uint, component string) ([]*TraceSampling, error)
	DeleteHostToTrace(apiID uint32, traceSourceID uint, component string) error
	DeleteAll() error
	ResetHostsToTrace(traceSourceID uint, component string) error
	GetExternalTraceSourceID() (uint, error)
}

type TraceSamplingTableHandler struct {
	tx *gorm.DB
}

func (h *TraceSamplingTableHandler) AddHostToTrace(apiID uint32, traceSourceID uint, component string) error {
	sampling := TraceSampling{
		APIID:         uint(apiID),
		TraceSourceID: traceSourceID,
		Component:     component,
	}
	return h.tx.Where(sampling).FirstOrCreate(sampling).Error
}

func (h *TraceSamplingTableHandler) GetHostsToTrace(traceSourceID uint, component string) ([]*TraceSampling, error) {
	var samplings []*TraceSampling
	t := h.tx.Where("trace_source_id = ? AND component = ?", traceSourceID, component)

	if err := t.Find(&samplings).Error; err != nil {
		return nil, err
	}

	return samplings, nil
}

func (h *TraceSamplingTableHandler) DeleteHostToTrace(apiID uint32, traceSourceID uint, component string) error {
	return h.tx.Unscoped().Delete(&TraceSampling{
		APIID:         uint(apiID),
		TraceSourceID: traceSourceID,
		Component:     component,
	}).Error
}

func (h *TraceSamplingTableHandler) DeleteAll() error {
	return h.tx.Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&TraceSampling{}).
		Error
}

func (h *TraceSamplingTableHandler) ResetHostsToTrace(traceSourceID uint, component string) error {
	return h.tx.Where("trace_source_id = ? AND component = ?", traceSourceID, component).
		Delete(&TraceSampling{}).
		Error
}

func (h *TraceSamplingTableHandler) GetExternalTraceSourceID() (uint, error) {
	return externalTraceSourceID, nil
}
