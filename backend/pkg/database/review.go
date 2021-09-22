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
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	reviewTableName = "reviews"

	// NOTE: when changing one of the column names change also the gorm label in APIEvent.
	approvedColumnName = "approved"
)

type Review struct {
	// will be populated after inserting to DB
	ID uint `gorm:"primarykey" faker:"-"`
	// CreatedAt time.Time
	// UpdatedAt time.Time

	Approved bool   `json:"approved,omitempty" gorm:"column:approved" faker:"-"`
	SpecKey  string `json:"specKey,omitempty" gorm:"column:speckey" faker:"-"`
	// serialized PathToPathItem from Speculator
	PathToPathItemStr string `json:"pathToPathItemStr,omitempty" gorm:"column:pathtopathitemstr" faker:"-"`
}

func (Review) TableName() string {
	return reviewTableName
}

func GetReviewTable() *gorm.DB {
	return DB.Table(reviewTableName)
}

func CreateReview(review *Review) error {
	if err := GetReviewTable().Create(&review).Error; err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func UpdateApprovedReview(approved bool, id uint32) error {
	if err := GetReviewTable().Model(&Review{}).Where("id = ?", id).Updates(map[string]interface{}{approvedColumnName: approved}).Error; err != nil {
		return err
	}

	return nil
}
