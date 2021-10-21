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
	"github.com/apiclarity/apiclarity/api/server/models"
	"github.com/apiclarity/apiclarity/api/server/restapi/operations"
	speculatorspec "github.com/apiclarity/speculator/pkg/spec"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbNameEnvVar     = "DB_NAME"
	DBUserEnvVar     = "DB_USER"
	DBPasswordEnvVar = "DB_PASS"
	DBHostEnvVar     = "DB_HOST"
	DBPortEnvVar     = "DB_PORT_NUMBER"
	FakeDataEnvVar   = "FAKE_DATA"
	FakeTracesEnvVar = "FAKE_TRACES"
	FakeDBPath       = "./db.db"
	enableDBInfoLogs = "ENABLE_DB_INFO_LOGS"
)

type Database interface {
	GetDB() *gorm.DB
	// apiEvent
	GroupByAPIInfo() ([]HostGroup, error)
	GetAPIEventsAndTotal(params operations.GetAPIEventsParams) ([]APIEvent, int64, error)
	GetAPIEvent(eventID uint32) (*APIEvent, error)
	GetAPIEventReconstructedSpecDiff(eventID uint32) (*APIEvent, error)
	GetAPIEventProvidedSpecDiff(eventID uint32) (*APIEvent, error)
	SetAPIEventsReconstructedPathID(approvedReview []*speculatorspec.ApprovedSpecReviewPathItem, host string, port string) error
	GetAPIEventsLatestDiffs(latestDiffsNum int) ([]APIEvent, error)
	GetAPIUsages(params operations.GetAPIUsageHitCountParams) ([]*models.HitCount, error)
	GetDashboardAPIUsages(startTime, endTime time.Time, apiType APIUsageType) ([]*models.APIUsage, error)

	// apiInventory
	GetAPIInventoryAndTotal(params operations.GetAPIInventoryParams) ([]APIInfo, int64, error)
	GetAPISpecs(apiID uint32) (*APIInfo, error)
	GetAPISpecsInfo(apiID uint32) (*models.OpenAPISpecs, error)
	PutAPISpec(apiID uint, spec string, specInfo *models.SpecInfo, specType specType) error
	DeleteProvidedAPISpec(apiID uint32) error
	DeleteApprovedAPISpec(apiID uint32) error
	GetAPIID(name, port string) (uint, error)
	GetAPIInventoryTableFirst(dest *APIInfo, conds ...interface{}) error

	// review
	UpdateApprovedReview(approved bool, id uint32) error
	CreateReview(review *Review) error
	GetReviewTableFirst(dest *Review, conds ...interface{}) error
}

type DatabaseHandler struct {
	DB *gorm.DB
}

func (db *DatabaseHandler) GetDB() *gorm.DB {
	return db.DB
}

func Init() *DatabaseHandler {
	databaseHandler := DatabaseHandler{}

	viper.AutomaticEnv()
	if viper.GetBool(FakeDataEnvVar) || viper.GetBool(FakeTracesEnvVar) {
		cleanFakeDataBase(FakeDBPath)
		databaseHandler.DB = initFakeDataBase(FakeDBPath)
	} else {
		databaseHandler.DB = initDataBase()
	}
	return &databaseHandler
}

func cleanFakeDataBase(databasePath string) {
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		log.Debug("deleting db...")
		if err := os.Remove(databasePath); err != nil {
			log.Warnf("failed to delete db file (%v): %v", databasePath, err)
		}
	}
}

func initDataBase() *gorm.DB {
	dbPass := viper.GetString(DBPasswordEnvVar)
	dbUser := viper.GetString(DBUserEnvVar)
	dbHost := viper.GetString(DBHostEnvVar)
	dbPort := viper.GetString(DBPortEnvVar)
	dbName := viper.GetString(dbNameEnvVar)

	dbLogger := logger.Default
	if viper.GetBool(enableDBInfoLogs) {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open %s db: %v", dbName, err)
	}

	// this will ensure table is created
	if err := db.AutoMigrate(&APIEvent{}, &APIInfo{}, &Review{}); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	return db
}

func initFakeDataBase(databasePath string) *gorm.DB {
	dbLogger := logger.Default
	if viper.GetBool(enableDBInfoLogs) {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	temp, _ := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: dbLogger,
	})
	// this will ensure table is created
	if err := temp.AutoMigrate(&APIEvent{}, &APIInfo{}, &Review{}); err != nil {
		panic(err)
	}

	return temp
}
