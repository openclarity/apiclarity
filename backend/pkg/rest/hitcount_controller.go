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

package rest

import (
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
)

const hitCountGranularity = 50

func (s *RESTServer) GetAPIUsageHitCount(params operations.GetAPIUsageHitCountParams) middleware.Responder {
	hitCounts, err := getAPIUsages(params)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetAPIUsageHitCountDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	return operations.NewGetAPIUsageHitCountOK().WithPayload(hitCounts)
}

func getAPIUsages(params operations.GetAPIUsageHitCountParams) ([]*models.HitCount, error) {
	var apiUsages []*models.HitCount

	startTime := time.Time(params.StartTime)
	endTime := time.Time(params.EndTime)
	diff := endTime.Sub(startTime)
	timeInterval := diff / hitCountGranularity

	db := database.SetAPIEventsFilters(database.GetAPIEventsTable(), getAPIUsageHitCountParamsToFilters(params), false).
		Session(&gorm.Session{})

	for i := 0; i < hitCountGranularity; i++ {
		var count int64
		st := strfmt.DateTime(startTime)
		et := strfmt.DateTime(startTime.Add(timeInterval))

		if err := db.Where(database.CreateTimeFilter(st, et)).Count(&count).Error; err != nil {
			return nil, err
		}

		apiUsages = append(apiUsages, &models.HitCount{
			Count: count,
			Time:  st,
		})

		startTime = startTime.Add(timeInterval)
	}
	return apiUsages, nil
}

func getAPIUsageHitCountParamsToFilters(params operations.GetAPIUsageHitCountParams) *database.APIEventsFilters {
	return &database.APIEventsFilters{
		DestinationIPIsNot:   params.DestinationIPIsNot,
		DestinationIPIs:      params.DestinationIPIs,
		DestinationPortIsNot: params.DestinationPortIsNot,
		DestinationPortIs:    params.DestinationPortIs,
		EndTime:              params.EndTime,
		ShowNonAPI:           params.ShowNonAPI,
		HasSpecDiffIs:        params.HasSpecDiffIs,
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
		StartTime:            params.StartTime,
		StatusCodeGte:        params.StatusCodeGte,
		StatusCodeIsNot:      params.StatusCodeIsNot,
		StatusCodeIs:         params.StatusCodeIs,
		StatusCodeLte:        params.StatusCodeLte,
	}
}
