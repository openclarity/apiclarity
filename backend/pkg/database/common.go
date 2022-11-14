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
	"strings"

	"github.com/go-openapi/strfmt"
	"gorm.io/gorm"

	"github.com/openclarity/apiclarity/api/server/models"
)

func FieldInTable(table, field string) string {
	return table + "." + field
}

func Paginate(page, pageSize int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * pageSize
		return db.Offset(int(offset)).Limit(int(pageSize))
	}
}

func CreateTimeFilter(columnName string, startTime, endTime strfmt.DateTime) string {
	return fmt.Sprintf("%s BETWEEN '%v' AND '%v'", columnName, startTime, endTime)
}

func CreateSortOrder(sortKey string, sortDir *string) (string, error) {
	sortKeyColumnName, err := getSortKeyColumnName(sortKey)
	if err != nil {
		return "", fmt.Errorf("failed to get sort key column name: %v", err)
	}

	return fmt.Sprintf("%v %v", sortKeyColumnName, strings.ToLower(*sortDir)), nil
}

func getSortKeyColumnName(key string) (string, error) {
	switch models.APIEventSortKey(key) {
	case models.APIEventSortKeyTime:
		return timeColumnName, nil
	case models.APIEventSortKeyMethod:
		return methodColumnName, nil
	case models.APIEventSortKeyPath:
		return pathColumnName, nil
	case models.APIEventSortKeyStatusCode:
		return statusCodeColumnName, nil
	case models.APIEventSortKeySourceIP:
		return sourceIPColumnName, nil
	case models.APIEventSortKeyDestinationIP:
		return destinationIPColumnName, nil
	case models.APIEventSortKeyDestinationPort:
		return destinationPortColumnName, nil
	case models.APIEventSortKeySpecDiffType:
		return specDiffTypeColumnName, nil
	case models.APIEventSortKeyHostSpecName:
		return hostSpecNameColumnName, nil
	case models.APIEventSortKeyAPIType:
		return eventTypeColumnName, nil
	}

	switch models.APIInventorySortKey(key) {
	case models.APIInventorySortKeyName:
		return nameColumnName, nil
	case models.APIInventorySortKeyPort:
		return portColumnName, nil
	case models.APIInventorySortKeyHasReconstructedSpec:
		return hasReconstructedSpecColumnName, nil
	case models.APIInventorySortKeyHasProvidedSpec:
		return hasProvidedSpecColumnName, nil
	}

	return "", fmt.Errorf("unknown sort key (%v)", key)
}

func FilterIsBool(db *gorm.DB, column string, value *bool) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s = ?", column), *value)
}

func FilterIsUint32(db *gorm.DB, column string, value *uint32) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s = ?", column), *value)
}

func FilterIs(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s IN ?", column), values)
}

func FilterIsOrNull(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("(%s IN ? OR %s IS NULL)", column, column), values)
}

func FilterIsNot(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	return db.Where(fmt.Sprintf("%s NOT IN ?", column), values)
}

func FilterStartsWith(db *gorm.DB, column string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	// ex. WHERE CustomerName LIKE 'a%'	Finds any values that start with "a"
	return db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%s%%", *value))
}

func FilterEndsWith(db *gorm.DB, column string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	// ex. WHERE CustomerName LIKE '%a'	Finds any values that end with "a"
	return db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%%%s", *value))
}

func FilterContains(db *gorm.DB, column string, values []string) *gorm.DB {
	if len(values) == 0 {
		return db
	}
	for _, value := range values {
		// ex. WHERE CustomerName LIKE '%or%'	Finds any values that have "or" in any position
		db = db.Where(fmt.Sprintf("%s LIKE ?", column), fmt.Sprintf("%%%s%%", value))
	}
	return db
}

func FilterGte(db *gorm.DB, column string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s >= ?", column), value)
}

func FilterLte(db *gorm.DB, column string, value *string) *gorm.DB {
	if value == nil {
		return db
	}
	return db.Where(fmt.Sprintf("%s <= ?", column), value)
}
