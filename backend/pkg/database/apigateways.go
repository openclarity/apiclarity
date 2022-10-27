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
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	apiGatewaysTableName = "api_gateways"
)

const (
	tokenByteLength = 32
)

type APIGateway struct {
	gorm.Model

	Name        string `json:"name,omitempty" gorm:"column:name;uniqueIndex" faker:"oneof: customer1.apigee.gw, mynicegateway"`
	Type        string `json:"type,omitempty" gorm:"column:type" faker:"oneof: KONG, TYK, APIGEEX"`
	Description string `json:"description,omitempty" gorm:"column:description" faker:"-"`
	Token       []byte `json:"auth_token,omitempty" gorm:"column:auth_token" faker:"-"`
}

type APIGatewaysTable interface {
	CreateAPIGateway(gateway *APIGateway) error
	GetAPIGateway(ID uint) (*APIGateway, error)
	GetAPIGateways() ([]*APIGateway, error)
	DeleteAPIGateway(ID uint) error
}

type APIGatewaysTableHandler struct {
	tx *gorm.DB
}

func (h *APIGatewaysTableHandler) CreateAPIGateway(gateway *APIGateway) error {
	return h.tx.Where(*gateway).FirstOrCreate(gateway).Error
}

func (gw *APIGateway) BeforeCreate(tx *gorm.DB) error {
	gw.Token = make([]byte, tokenByteLength)
	if _, err := rand.Read(gw.Token); err != nil {
		log.Errorf("Unable to generate token for APIGateway '%d': %v", gw.ID, err)
		return fmt.Errorf("unable to generate token for APIGateway '%d': %v", gw.ID, err)
	}

	return nil
}

func (h *APIGatewaysTableHandler) GetAPIGateway(ID uint) (*APIGateway, error) {
	egw := APIGateway{}
	if err := h.tx.First(&egw, ID).Error; err != nil {
		return nil, err
	}

	return &egw, nil
}

func (h *APIGatewaysTableHandler) GetAPIGateways() ([]*APIGateway, error) {
	dest := []*APIGateway{}

	h.tx.Find(&dest)
	return dest, nil
}

func (h *APIGatewaysTableHandler) DeleteAPIGateway(ID uint) error {
	return h.tx.Unscoped().Delete(&APIGateway{}, ID).Error
}
