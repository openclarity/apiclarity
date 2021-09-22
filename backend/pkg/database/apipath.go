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

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	apiPathTableName = "api_paths"

	// NOTE: when changing one of the column names change also the gorm label in APIPath.
	apiPathPathIDColumnName = "id"
	apiPathPathColumnName   = "path"
)

type APIPath struct {
	ID string `gorm:"primarykey" faker:"-"`

	// Path as shown in specs, might be parametrized
	Path string `json:"path,omitempty" gorm:"column:path" faker:"-"`
	// APIID
	APIID uint `json:"apiId,omitempty" gorm:"column:apiid" faker:"-"`
}

func (APIPath) TableName() string {
	return apiPathTableName
}

func GetAPIPathsTable() *gorm.DB {
	return DB.Table(apiPathTableName)
}

func CreateAPIPath(path *APIPath) error {
	if result := GetAPIPathsTable().Create(path); result.Error != nil {
		return fmt.Errorf("failed to create api path: %v", result.Error)
	}

	log.Infof("API path was created: %+v", path)

	return nil
}

func StorePaths(paths []*APIPath) {
	for _, path := range paths {
		if err := CreateAPIPath(path); err != nil {
			log.Warnf("Failed to store path (%v): %v", path.Path, err)
		}
	}
}

func GetPathIDs(path string) ([]string, error) {
	var pathIds []string
	if result := GetAPIPathsTable().Select(apiPathPathIDColumnName).Where(apiPathPathColumnName+" = ?", path).Find(&pathIds); result.Error != nil {
		return nil, result.Error
	}

	return pathIds, nil
}
