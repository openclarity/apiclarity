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
	"fmt"
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
	LocalDBPath      = "./db.db"
	enableDBInfoLogs = "ENABLE_DB_INFO_LOGS"
)

const (
	DBDriverTypePostgres = "POSTGRES"
	DBDriverTypeLocal    = "LOCAL"
)

type Database interface {
	APIEventsTable() APIEventsTable
	APIInventoryTable() APIInventoryTable
	ReviewTable() ReviewTable
}

type Handler struct {
	DB *gorm.DB
}

type DBConfig struct {
	DriverType string
}

func Init(config *DBConfig) *Handler {
	databaseHandler := Handler{}

	databaseHandler.DB = initDataBase(config)

	return &databaseHandler
}

func (db *Handler) APIEventsTable() APIEventsTable {
	return &APIEventsTableHandler{
		tx: db.DB.Table(apiEventTableName),
	}
}

func (db *Handler) APIInventoryTable() APIInventoryTable {
	return &APIInventoryTableHandler{
		tx: db.DB.Table(apiInventoryTableName),
	}
}

func (db *Handler) ReviewTable() ReviewTable {
	return &ReviewTableHandler{
		tx: db.DB.Table(reviewTableName),
	}
}

func cleanLocalDataBase(databasePath string) {
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		log.Debug("deleting db...")
		if err := os.Remove(databasePath); err != nil {
			log.Warnf("failed to delete db file (%v): %v", databasePath, err)
		}
	}
}

func initDataBase(config *DBConfig) *gorm.DB {
	var db *gorm.DB
	dbDriver := config.DriverType
	dbLogger := logger.Default
	if viper.GetBool(enableDBInfoLogs) {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	switch dbDriver {
	case DBDriverTypePostgres:
		db = initPostgres(dbLogger)
	case DBDriverTypeLocal:
		db = initSqlite(dbLogger)
	default:
		log.Fatalf("DB driver is not supported: %v", dbDriver)
	}

	// this will ensure table is created
	if err := db.AutoMigrate(&APIEvent{}, &APIInfo{}, &Review{}); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	return db
}

func initPostgres(dbLogger logger.Interface) *gorm.DB {
	dbPass := viper.GetString(DBPasswordEnvVar)
	dbUser := viper.GetString(DBUserEnvVar)
	dbHost := viper.GetString(DBHostEnvVar)
	dbPort := viper.GetString(DBPortEnvVar)
	dbName := viper.GetString(dbNameEnvVar)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPass, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open %s db: %v", dbName, err)
	}

	return db
}

func initSqlite(dbLogger logger.Interface) *gorm.DB {
	cleanLocalDataBase(LocalDBPath)

	db, err := gorm.Open(sqlite.Open(LocalDBPath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	return db
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
