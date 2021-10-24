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
	"gorm.io/gorm"

	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
)

type APIUsageType string

const (
	APIWithDiffs APIUsageType = "APIWithDiffs"
	ExistingAPI  APIUsageType = "ExistingAPI"
	NewAPI       APIUsageType = "NewAPI"
)

func (a *APIEventsTableHandler) getAPIUsageDBSession(apiType APIUsageType) (db *gorm.DB, err error) {
	switch apiType {
	case APIWithDiffs:
		db = a.tx.Where(hasSpecDiffColumnName+" = ?", true).Session(&gorm.Session{})
	case ExistingAPI:
		// REST api (not a non-api)
		// no spec diff
		// have reconstructed OR provided spec
		db = a.tx.
			Where(FieldInTable(apiEventTableName, isNonAPIColumnName)+" = ?", false).
			Where(FieldInTable(apiEventTableName, hasSpecDiffColumnName)+" = ?", false).
			Where(FieldInTable(apiInventoryTableName, hasReconstructedSpecColumnName)+" = ? OR "+
				FieldInTable(apiInventoryTableName, hasProvidedSpecColumnName)+" = ?", true, true).
			Joins("left join " + apiInventoryTableName + " on " + FieldInTable(apiInventoryTableName, idColumnName) +
				" = " + FieldInTable(apiEventTableName, apiInfoIDColumnName)).
			Session(&gorm.Session{})
	case NewAPI:
		// REST api (not a non-api)
		// no spec diff
		// no reconstructed AND no provided spec
		db = a.tx.
			Where(FieldInTable(apiEventTableName, isNonAPIColumnName)+" = ?", false).
			Where(FieldInTable(apiEventTableName, hasSpecDiffColumnName)+" = ?", false).
			Where(FieldInTable(apiInventoryTableName, hasReconstructedSpecColumnName)+" = ? AND "+
				FieldInTable(apiInventoryTableName, hasProvidedSpecColumnName)+" = ?", false, false).
			Joins("left join " + apiInventoryTableName + " on " + FieldInTable(apiInventoryTableName, idColumnName) +
				" = " + FieldInTable(apiEventTableName, apiInfoIDColumnName)).
			Session(&gorm.Session{})
	default:
		return nil, fmt.Errorf("unknown API type: %v", apiType)
	}

	return db, nil
}

func (a *APIEventsTableHandler) GetDashboardAPIUsages(startTime, endTime time.Time, apiType APIUsageType) ([]*models.APIUsage, error) {
	var apiUsages []*models.APIUsage
	var count int64

	diff := endTime.Sub(startTime)

	timeInterval := diff / hitCountGranularity

	db, err := a.getAPIUsageDBSession(apiType)
	if err != nil {
		return nil, fmt.Errorf("failed to get DB session: %v", err)
	}

	for i := 0; i < hitCountGranularity; i++ {
		endTime := startTime.Add(timeInterval)
		st := strfmt.DateTime(startTime)
		et := strfmt.DateTime(endTime)

		if err := db.Where(CreateTimeFilter(st, et)).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to query DB: %v", err)
		}

		apiUsages = append(apiUsages, &models.APIUsage{
			Time:       st,
			NumOfCalls: count,
		})

		startTime = endTime
	}
	return apiUsages, nil
}

const hitCountGranularity = 50

func (a *APIEventsTableHandler) GetAPIUsages(params operations.GetAPIUsageHitCountParams) ([]*models.HitCount, error) {
	var apiUsages []*models.HitCount

	startTime := time.Time(params.StartTime)
	endTime := time.Time(params.EndTime)
	diff := endTime.Sub(startTime)
	timeInterval := diff / hitCountGranularity

	db := a.setAPIEventsFilters(getAPIUsageHitCountParamsToFilters(params), false).
		Session(&gorm.Session{})

	for i := 0; i < hitCountGranularity; i++ {
		var count int64
		st := strfmt.DateTime(startTime)
		et := strfmt.DateTime(startTime.Add(timeInterval))

		if err := db.Where(CreateTimeFilter(st, et)).Count(&count).Error; err != nil {
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
