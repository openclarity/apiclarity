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

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	"github.com/apiclarity/apiclarity/backend/pkg/database"
)

func (s *RESTServer) GetDashboardAPIUsage(params operations.GetDashboardAPIUsageParams) middleware.Responder {
	var apisWithDiff []*models.APIUsage
	var existingApis []*models.APIUsage
	var newApis []*models.APIUsage

	apisWithDiffUsage, err := getDashboardAPIUsages(time.Time(params.StartTime), time.Time(params.EndTime), database.ApiWithDiffs)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetDashboardAPIUsageDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}
	for _, usage := range apisWithDiffUsage {
		apisWithDiff = append(apisWithDiff, usage)
	}

	existingApisUsage, err := getDashboardAPIUsages(time.Time(params.StartTime), time.Time(params.EndTime), database.ExistingApi)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetDashboardAPIUsageDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}
	for _, usage := range existingApisUsage {
		existingApis = append(existingApis, usage)
	}

	newApisUsage, err := getDashboardAPIUsages(time.Time(params.StartTime), time.Time(params.EndTime), database.NewApi)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetDashboardAPIUsageDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	for _, usage := range newApisUsage {
		newApis = append(newApis, usage)
	}

	return operations.NewGetDashboardAPIUsageOK().WithPayload(&models.APIUsages{
		ApisWithDiff: apisWithDiff,
		ExistingApis: existingApis,
		NewApis:      newApis,
	})
}

const latestDiffsNum = 5

func (s *RESTServer) GetDashboardAPIUsageLatestDiffs(params operations.GetDashboardAPIUsageLatestDiffsParams) middleware.Responder {
	var diffs []*models.SpecDiffTime

	latestDiffs, err := database.GetAPIEventsLatestDiffs(latestDiffsNum)
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetDashboardAPIUsageLatestDiffsDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}

	for _, diff := range latestDiffs {
		diffs = append(diffs, &models.SpecDiffTime{
			APIHostName: diff.HostSpecName,
			APIEventID:  uint32(diff.ID),
			Time:        diff.Time,
		})
	}

	return operations.NewGetDashboardAPIUsageLatestDiffsOK().WithPayload(diffs)
}

func (s *RESTServer) GetDashboardAPIUsageMostUsed(_ operations.GetDashboardAPIUsageMostUsedParams) middleware.Responder {

	var ret []*models.APICount

	groups, err := database.GroupByAPIInfo(database.GetAPIEventsTable())
	if err != nil {
		// TODO: need to handle errors
		// https://github.com/go-gorm/gorm/blob/master/errors.go
		log.Error(err)
		return operations.NewGetDashboardAPIUsageMostUsedDefault(500).WithPayload(&models.APIResponse{
			Message: "Oops",
		})
	}
	for _, group := range groups {
		ret = append(ret, &models.APICount{
			APIHostName: group.HostSpecName,
			APIPort:     group.Port,
			APIType:     models.APIType(group.ApiType),
			APIInfoID:   group.ApiInfoId,
			NumCalls:    int64(group.Count),
		})
	}

	return operations.NewGetDashboardAPIUsageMostUsedOK().WithPayload(ret)
}

func getDashboardAPIUsages(startTime, endTime time.Time, apiType database.ApiUsageType) ([]*models.APIUsage, error) {
	var apiUsages []*models.APIUsage
	var count int64

	diff := endTime.Sub(startTime)

	timeInterval := diff / hitCountGranularity

	db, err := database.GetAPIUsageDBSession(apiType)
	if err != nil {
		return nil, err
	}

	for i := 0; i < hitCountGranularity; i++ {
		endTime := startTime.Add(timeInterval)
		st := strfmt.DateTime(startTime)
		et := strfmt.DateTime(endTime)

		if err := db.Where(database.CreateTimeFilter(st, et)).Count(&count).Error; err != nil {
			return nil, err
		}

		apiUsages = append(apiUsages, &models.APIUsage{
			Time:       st,
			NumOfCalls: count,
		})

		startTime = endTime
	}
	return apiUsages, nil
}
