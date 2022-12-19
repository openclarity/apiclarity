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
	"strconv"

	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/backend/pkg/common"
	"github.com/openclarity/apiclarity/backend/pkg/utils"
)

const (
	traceSamplingTableName = "trace_sampling"
	hostnamePortSeparator  = ":"
)

type TraceSampling struct {
	gorm.Model

	APIID         uint   `json:"api_id,omitempty" gorm:"column:api_id" faker:"-"`
	TraceSourceID uint   `json:"trace_source_id,omitempty" gorm:"column:trace_source_id" faker:"-"`
	Component     string `json:"component,omitempty" gorm:"column:component" faker:"-"`

	APIInfo     APIInfo     `gorm:"foreignKey:APIID;constraint:OnDelete:CASCADE"`
	TraceSource TraceSource `gorm:"constraint:OnDelete:CASCADE"`
}

type TraceSamplingWithHostAndPort struct {
	APIID                uint
	TraceSourceID        uint
	Component            string
	Name                 string
	Port                 uint
	DestinationNamespace string
}

type TraceSamplingTable interface {
	AddAPIToTrace(component string, traceSourceID uint, apiID uint32) error
	GetAPIsToTrace(component string, traceSourceID uint) ([]*TraceSampling, error)
	DeleteAPIToTrace(component string, traceSourceID uint, apiID uint32) error
	DeleteAll() error
	ResetAPIsToTraceByTraceSource(component string, traceSourceID uint) error
	ResetAPIsToTraceByComponent(component string) error
	GetExternalTraceSourceID() (uint, error)
	HostsToTraceByTraceSource(component string, traceSourceID uint) ([]string, error)
	HostsToTraceByComponent(component string) (map[uint][]string, error)
}

type TraceSamplingTableHandler struct {
	tx *gorm.DB
}

func (TraceSampling) TableName() string {
	return traceSamplingTableName
}

func (h *TraceSamplingTableHandler) AddAPIToTrace(component string, traceSourceID uint, apiID uint32) error {
	sampling := TraceSampling{
		APIID:         uint(apiID),
		TraceSourceID: traceSourceID,
		Component:     component,
	}
	return h.tx.Where(sampling).FirstOrCreate(&sampling).Error
}

func (h *TraceSamplingTableHandler) GetAPIsToTrace(component string, traceSourceID uint) ([]*TraceSampling, error) {
	var samplings []*TraceSampling
	t := h.tx.Where("trace_source_id = ? AND component = ?", traceSourceID, component)

	if err := t.Find(&samplings).Error; err != nil {
		return nil, err
	}

	return samplings, nil
}

func (h *TraceSamplingTableHandler) DeleteAPIToTrace(component string, traceSourceID uint, apiID uint32) error {
	return h.tx.Unscoped().
		Where("trace_source_id = ? AND component = ? AND api_id = ?", traceSourceID, component, apiID).
		Delete(&TraceSampling{}).
		Error
}

func (h *TraceSamplingTableHandler) DeleteAll() error {
	return h.tx.Session(&gorm.Session{AllowGlobalUpdate: true}).
		Delete(&TraceSampling{}).
		Error
}

func (h *TraceSamplingTableHandler) ResetAPIsToTraceByTraceSource(component string, traceSourceID uint) error {
	return h.tx.Where("trace_source_id = ? AND component = ?", traceSourceID, component).
		Delete(&TraceSampling{}).
		Error
}

func (h *TraceSamplingTableHandler) ResetAPIsToTraceByComponent(component string) error {
	return h.tx.Where("component = ?", component).
		Delete(&TraceSampling{}).
		Error
}

func (h *TraceSamplingTableHandler) GetExternalTraceSourceID() (uint, error) {
	return common.DefaultTraceSourceID, nil
}

// createHostFromTraceSamplingWithHostAndPort will create hosts in the format of `hostname:port` if port exist, otherwise will return only hostname
// Note: The function will return both `hostname:port` and `hostname` in case port is the default HTTP port (80).
func createHostFromTraceSamplingWithHostAndPort(sampling *TraceSamplingWithHostAndPort) (ret []string) {
	// TODO: we might need to create multiple hosts from a single api.Host
	// example: hostname=foo, port=8080 ==> host=[foo:8080, foo.namespace:8080, ....]
	if sampling.Port > 0 {
		ret = append(ret, sampling.Name+hostnamePortSeparator+strconv.Itoa(int(sampling.Port)))
	}

	if sampling.Port == 0 || sampling.Port == 80 {
		ret = append(ret, sampling.Name)
	}

	return ret
}

func (h *TraceSamplingTableHandler) HostsToTraceByTraceSource(component string, traceSourceID uint) ([]string, error) {
	var hosts []string

	var samplings []*TraceSamplingWithHostAndPort
	var t *gorm.DB
	if component == "*" {
		t = h.tx.Select("trace_sampling.api_id, trace_sampling.trace_source_id, trace_sampling.component, api_inventory.name, api_inventory.port, api_inventory.destination_namespace").
			Where("trace_sampling.trace_source_id = ?", traceSourceID).
			Joins("LEFT JOIN api_inventory ON api_inventory.id = trace_sampling.api_id")
	} else {
		t = h.tx.Select("trace_sampling.api_id, trace_sampling.trace_source_id, trace_sampling.component, api_inventory.name, api_inventory.port, api_inventory.destination_namespace").
			Where("trace_sampling.trace_source_id = ? AND (trace_sampling.component = ? OR trace_sampling.component = ?)", traceSourceID, component, "*").
			Joins("LEFT JOIN api_inventory ON api_inventory.id = trace_sampling.api_id")
	}
	if err := t.Find(&samplings).Error; err != nil {
		return nil, err
	}

	for _, sampling := range samplings {
		hosts = append(hosts, createHostFromTraceSamplingWithHostAndPort(sampling)...)
	}

	// Remove as soon as possible duplicate hosts (because of component="*")
	hosts = utils.RemoveDuplicateStringFromSlice(hosts)

	return hosts, nil
}

func (h *TraceSamplingTableHandler) HostsToTraceByComponent(component string) (map[uint][]string, error) {
	hostsMap := make(map[uint][]string)

	var samplings []*TraceSamplingWithHostAndPort
	t := h.tx.Select("trace_sampling.api_id, trace_sampling.trace_source_id, trace_sampling.component, api_inventory.name, api_inventory.port, api_inventory.destination_namespace").
		Where("trace_sampling.component = ? OR trace_sampling.component = ?", component, "*").
		Joins("LEFT JOIN api_inventory ON api_inventory.id = trace_sampling.api_id")
	if err := t.Find(&samplings).Error; err != nil {
		return nil, err
	}
	for _, sampling := range samplings {
		traceSourceHosts := createHostFromTraceSamplingWithHostAndPort(sampling)
		hostsMap[sampling.TraceSourceID] = append(hostsMap[sampling.TraceSourceID], traceSourceHosts...)
	}

	// Remove as soon as possible duplicate hosts (because of component="*")
	for traceSourceID, hosts := range hostsMap {
		hostsMap[traceSourceID] = utils.RemoveDuplicateStringFromSlice(hosts)
	}

	return hostsMap, nil
}
