/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package database

import (
	"fmt"
	"gorm.io/gorm"
)

type ApiUsageType string

const (
	ApiWithDiffs ApiUsageType = "ApiWithDiffs"
	ExistingApi  ApiUsageType = "ExistingApi"
	NewApi       ApiUsageType = "NewApi"
)

func GetAPIUsageDBSession(apiType ApiUsageType) (db *gorm.DB, err error) {
	switch apiType {
	case ApiWithDiffs:
		db = GetAPIEventsTable().Where(hasSpecDiffColumnName+" = ?", true).Session(&gorm.Session{})
	case ExistingApi:
		// REST api (not a non-api)
		// no spec diff
		// have reconstructed OR provided spec
		db = GetAPIEventsTable().
			Where(FieldInTable(apiEventTableName, isNonApiColumnName)+" = ?", false).
			Where(FieldInTable(apiEventTableName, hasSpecDiffColumnName)+" = ?", false).
			Where(FieldInTable(apiInventoryTableName, hasReconstructedSpecColumnName)+" = ? OR "+
				FieldInTable(apiInventoryTableName, hasProvidedSpecColumnName)+" = ?", true, true).
			Joins("left join " + apiInventoryTableName + " on " + FieldInTable(apiInventoryTableName, idColumnName) +
				" = " + FieldInTable(apiEventTableName, apiInfoIdColumnName)).
			Session(&gorm.Session{})
	case NewApi:
		// REST api (not a non-api)
		// no spec diff
		// no reconstructed AND no provided spec
		db = GetAPIEventsTable().
			Where(FieldInTable(apiEventTableName, isNonApiColumnName)+" = ?", false).
			Where(FieldInTable(apiEventTableName, hasSpecDiffColumnName)+" = ?", false).
			Where(FieldInTable(apiInventoryTableName, hasReconstructedSpecColumnName)+" = ? AND "+
				FieldInTable(apiInventoryTableName, hasProvidedSpecColumnName)+" = ?", false, false).
			Joins("left join " + apiInventoryTableName + " on " + FieldInTable(apiInventoryTableName, idColumnName) +
				" = " + FieldInTable(apiEventTableName, apiInfoIdColumnName)).
			Session(&gorm.Session{})
	default:
		return nil, fmt.Errorf("unknown API type: %v", apiType)
	}

	return db, nil
}
