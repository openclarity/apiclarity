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
	"errors"
	"fmt"

	"github.com/google/uuid"
	lru "github.com/hashicorp/golang-lru" // Move to v2 to use generics when go version will be updated
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	traceSourcesTableName = "trace_sources"
)

type TraceSource struct {
	ID uint `gorm:"primarykey"`

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

type TraceSourceTokenCache struct {
	cache *lru.Cache
}

const traceSourceTokenCacheSize = 128

func NewTraceSourceTokenCache() *TraceSourceTokenCache {
	c, err := lru.New(traceSourceTokenCacheSize)
	if err != nil {
		log.Fatalf("Unable to initialize Token Cache: %v", err)
	}

	return &TraceSourceTokenCache{
		cache: c,
	}
}

func (c *TraceSourceTokenCache) Get(token string) (*TraceSource, bool) {
	if value, ok := c.cache.Get(token); ok {
		ts := value.(*TraceSource)
		return ts, ok
	}

	return nil, false
}

func (c *TraceSourceTokenCache) Add(token string, ts *TraceSource) {
	c.cache.Add(token, ts)
}

func (c *TraceSourceTokenCache) Remove(token string) {
	c.cache.Remove(token)
}

func (c *TraceSourceTokenCache) UpdateTraceSource(traceSource *TraceSource) {
	for token := range c.cache.Keys() {
		value, _ := c.cache.Get(token)
		ts := value.(*TraceSource)
		if ts.ID == traceSource.ID {
			c.cache.Remove(token)
			break
		}
	}
	c.cache.Add(traceSource.Token, traceSource)
}

var tokensCache = NewTraceSourceTokenCache()

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
		if token, err := GenerateTraceSourceToken(); err != nil {
			log.Errorf("Unable to generate token for Trace Source '%d': %v", source.ID, err)
			return fmt.Errorf("Unable to generate random token for Trace Source: '%d'", source.ID)
		} else {
			source.Token = token
		}
	}

	// If no uuid is provided, create one
	if source.UID == uuid.Nil {
		source.UID = uuid.New()
	}

	return nil
}

func (source *TraceSource) AfterCreate(tx *gorm.DB) error {
	tokensCache.Add(source.Token, source)
	return nil
}

func (source *TraceSource) AfterDelete(tx *gorm.DB) error {
	tokensCache.Remove(source.Token)
	return nil
}

func (source *TraceSource) AfterUpdate(tx *gorm.DB) error {
	tokensCache.UpdateTraceSource(source)
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
	if cachedTS, ok := tokensCache.Get(token); ok {
		if cachedTS == nil { // It means that this token was cached as invalid
			return nil, fmt.Errorf("invalid token")
		}
		return cachedTS, nil
	}

	source := TraceSource{}
	if err := h.tx.First(&source, TraceSource{Token: token}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			tokensCache.Add(token, nil) // Cache this token as invalid
		}
		return nil, err
	}

	tokensCache.Add(token, &source)

	return &source, nil
}

func (h *TraceSourcesTableHandler) GetTraceSources() ([]*TraceSource, error) {
	dest := []*TraceSource{}

	h.tx.Find(&dest)
	return dest, nil
}

func (h *TraceSourcesTableHandler) DeleteTraceSource(uid uuid.UUID) error {
	// We need the returning clause in order to have the struct filled in the
	// AfterDelete hook.
	return h.tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "auth_token"}}}).Where(&TraceSource{UID: uid}).Delete(&TraceSource{}).Error
}
