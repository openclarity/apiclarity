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
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	traceSourcesTableName = "trace_sources"
)

type TraceSource struct {
	gorm.Model

	UID         uuid.UUID `json:"uid,omitempty" gorm:"column:uid;uniqueIndex;type:uuid"`
	Name        string    `json:"name,omitempty" gorm:"column:name;uniqueIndex" faker:"oneof: customer1.apigee.gw, mynicegateway"`
	Type        string    `json:"type,omitempty" gorm:"column:type" faker:"oneof: KONG, TYK, APIGEEX"`
	Description string    `json:"description,omitempty" gorm:"column:description" faker:"-"`
	Token       string    `json:"auth_token,omitempty" gorm:"column:auth_token;uniqueIndex" faker:"-"`
}

type TraceSourcesTable interface {
	Prepopulate() error
	CreateTraceSource(source *TraceSource) error
	GetTraceSource(uuid.UUID) (*TraceSource, error)
	GetTraceSourceFromToken(token string) (*TraceSource, error)
	GetTraceSources() ([]*TraceSource, error)
	DeleteTraceSource(uuid.UUID) error
}

type TraceSourcesTableHandler struct {
	tx *gorm.DB
}

func (h *TraceSourcesTableHandler) Prepopulate() error {
	defaultTraceSources := []map[string]interface{}{
		{"ID": 0, "Name": "Default Trace Source", "UID": uuid.Nil},
	}

	return h.tx.Model(&TraceSource{}).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&defaultTraceSources).Error
}

const tokenByteLength = 32

func GenerateTraceSourceToken() (string, error) {
	secret := make([]byte, tokenByteLength)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("unable to generate random secret: %v", err)
	}

	encoded := base64.URLEncoding.EncodeToString(secret)

	return encoded, nil
}

func (h *TraceSourcesTableHandler) CreateTraceSource(source *TraceSource) error {
	return h.tx.Where(*source).FirstOrCreate(source).Error
}

func (source *TraceSource) BeforeCreate(tx *gorm.DB) error {
	// If no token is provided, create one
	if len(source.Token) == 0 {
		newToken, err := GenerateTraceSourceToken()
		if err != nil {
			log.Errorf("Unable to generate token for Trace Source '%d': %v", source.ID, err)
			return fmt.Errorf("unable to generate random token for Trace Source: '%d'", source.ID)
		}
		source.Token = newToken
	}

	// If no uuid is provided, create one
	if source.UID == uuid.Nil {
		source.UID = uuid.New()
	}

	return nil
}

func (h *TraceSourcesTableHandler) GetTraceSource(uid uuid.UUID) (*TraceSource, error) {
	source := TraceSource{UID: uid}
	if err := h.tx.First(&source, source).Error; err != nil {
		return nil, err
	}

	return &source, nil
}

func (h *TraceSourcesTableHandler) GetTraceSourceFromToken(token string) (*TraceSource, error) {
	source := TraceSource{}
	if err := h.tx.First(&source, TraceSource{Token: token}).Error; err != nil {
		return nil, err
	}

	return &source, nil
}

func (h *TraceSourcesTableHandler) GetTraceSources() ([]*TraceSource, error) {
	dest := []*TraceSource{}

	h.tx.Find(&dest)
	return dest, nil
}

func (h *TraceSourcesTableHandler) DeleteTraceSource(uid uuid.UUID) error {
	return h.tx.Unscoped().Delete(&TraceSource{}, &TraceSource{UID: uid}).Error
}
