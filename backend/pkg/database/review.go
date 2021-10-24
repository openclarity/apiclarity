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
	"context"
	"time"

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
	SpecKey  string `json:"specKey,omitempty" gorm:"column:spec_key" faker:"-"`
	// serialized PathToPathItem from Speculator
	PathToPathItemStr string `json:"pathToPathItemStr,omitempty" gorm:"column:path_to_path_item_str" faker:"-"`
}

type ReviewTable interface {
	UpdateApprovedReview(approved bool, id uint32) error
	Create(review *Review) error
	First(dest *Review, conds ...interface{}) error
	DeleteApproved() error
}

type ReviewTableHandler struct {
	tx *gorm.DB
}

func (Review) TableName() string {
	return reviewTableName
}

func (r *ReviewTableHandler) Create(review *Review) error {
	if err := r.tx.Create(&review).Error; err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (r *ReviewTableHandler) UpdateApprovedReview(approved bool, id uint32) error {
	if err := r.tx.Model(&Review{}).Where("id = ?", id).Updates(map[string]interface{}{approvedColumnName: approved}).Error; err != nil {
		return err
	}

	return nil
}

func (r *ReviewTableHandler) DeleteApproved() error {
	return r.tx.Where("approved =  ?", true).Delete(Review{}).Error
}

func (r *ReviewTableHandler) First(dest *Review, conds ...interface{}) error {
	return r.tx.First(dest, conds).Error
}

func (db *Handler) StartReviewTableCleaner(ctx context.Context, cleanInterval time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Debugf("Stopping database cleaner")
				return
			case <-time.After(cleanInterval):
				if err := db.ReviewTable().DeleteApproved(); err != nil {
					log.Errorf("Failed to delete approved review from database. %v", err)
				}
			}
		}
	}()
}
