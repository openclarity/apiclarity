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
	"os"

	log "github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	FakeDataEnvVar   = "FAKE_DATA"
	FakeTracesEnvVar = "FAKE_TRACES"
	localDBPath      = "./db.db"
)

const (
	DBDriverTypePostgres = "POSTGRES"
	DBDriverTypeLocal    = "LOCAL"
)

//go:generate $GOPATH/bin/mockgen -destination=./mock_database.go -package=database github.com/openclarity/apiclarity/backend/pkg/database Database
type Database interface {
	APIEventsTable() APIEventsTable
	APIInventoryTable() APIInventoryTable
	ReviewTable() ReviewTable
	APIEventsAnnotationsTable() APIEventAnnotationTable
	APIInfoAnnotationsTable() APIAnnotationsTable
	TraceSourcesTable() TraceSourcesTable
	TraceSamplingTable() TraceSamplingTable
}

type Handler struct {
	DB *gorm.DB
}

type DBConfig struct {
	EnableInfoLogs bool
	DriverType     string
	DBPassword     string
	DBUser         string
	DBHost         string
	DBPort         string
	DBName         string
}

func Init(config *DBConfig) *Handler {
	databaseHandler := Handler{}

	databaseHandler.DB = initDataBase(config)

	if err := databaseHandler.TraceSourcesTable().Prepopulate(); err != nil {
		log.Fatalf("Unable to prepopulate TraceSource table: %v", err)
	}

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

func (db *Handler) APIEventsAnnotationsTable() APIEventAnnotationTable {
	return &APIEventAnnotationTableHandler{
		tx: db.DB.Table(eventAnnotationsTableName),
	}
}

func (db *Handler) APIInfoAnnotationsTable() APIAnnotationsTable {
	return &APIInfoAnnotationsTableHandler{
		tx: db.DB.Table(apiEventAnnotationsTableName),
	}
}

func (db *Handler) TraceSourcesTable() TraceSourcesTable {
	return &TraceSourcesTableHandler{
		tx: db.DB.Table(traceSourcesTableName),
	}
}

func (db *Handler) TraceSamplingTable() TraceSamplingTable {
	return &TraceSamplingTableHandler{
		tx: db.DB.Table(traceSamplingTableName),
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
	if config.EnableInfoLogs {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	switch dbDriver {
	case DBDriverTypePostgres:
		db = initPostgres(config, dbLogger)
	case DBDriverTypeLocal:
		db = initSqlite(dbLogger)
	default:
		log.Fatalf("DB driver is not supported: %v", dbDriver)
	}

	// this will ensure table is created
	if err := db.AutoMigrate(&APIEvent{},
		&APIInfo{},
		&Review{},
		&APIEventAnnotation{},
		&APIInfoAnnotation{},
		&TraceSource{},
		&TraceSampling{}); err != nil {
		log.Fatalf("Failed to run auto migration: %v", err)
	}

	return db
}

func initPostgres(config *DBConfig, dbLogger logger.Interface) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open %s db: %v", config.DBName, err)
	}

	return db
}

func initSqlite(dbLogger logger.Interface) *gorm.DB {
	cleanLocalDataBase(localDBPath)

	db, err := gorm.Open(sqlite.Open(localDBPath), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}

	// https://www.sqlite.org/foreignkeys.html#fk_enable
	db.Exec("PRAGMA foreign_keys = ON")

	return db
}
