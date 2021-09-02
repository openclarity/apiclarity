/*
 *
 * Copyright (c) 2020 Cisco Systems, Inc. and its affiliates.
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	dbName           = "apiclarity"
	DBUser           = "root" // TODO: We shouldn't use the root user
	DBPasswordEnvVar = "DB_PASS"
	DBHost           = "mysql"
	DBPort           = "3306"
	FakeDataEnvVar   = "FAKE_DATA"
	FakeTracesEnvVar = "FAKE_TRACES"
	FakeDBPath       = "./db.db"
	enableDbInfoLogs = "ENABLE_DB_INFO_LOGS"
)

var (
	DB *gorm.DB
)

func init() {
	viper.AutomaticEnv()
	if viper.GetBool(FakeDataEnvVar) || viper.GetBool(FakeTracesEnvVar) {
		cleanFakeDataBase(FakeDBPath)
		DB = initFakeDataBase(FakeDBPath)
	} else {
		DB = initDataBase()
	}
}

func cleanFakeDataBase(databasePath string) {
	if _, err := os.Stat(databasePath); !os.IsNotExist(err) {
		log.Infof("deleting db...")
		if err := os.Remove(databasePath); err != nil {
			log.Warnf("failed to delete db file (%v): %v", databasePath, err)
		}
	}
}

func initDataBase() *gorm.DB {
	pass := viper.GetString(DBPasswordEnvVar)
	// The env var for some reason has new line at the end
	pass = strings.TrimRight(pass, "\n")

	sqldb, err := sql.Open("mysql", DBUser+":"+pass+"@tcp("+DBHost+":"+DBPort+")/")
	if err != nil {
		panic(err)
	}
	_, err = sqldb.Exec("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		panic(err)
	}

	if err := sqldb.Close(); err != nil {
		log.Errorf("Failed to close the initial mysql connection")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DBUser, pass, DBHost, DBPort, dbName)

	dbLogger := logger.Default
	if viper.GetBool(enableDbInfoLogs) {
		dbLogger = dbLogger.LogMode(logger.Info)
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		panic(err.Error())
	}
	// this will ensure table is created
	if err := db.AutoMigrate(&APIEvent{}, &APIInfo{}, &Review{}, &APIPath{}); err != nil {
		panic(err)
	}

	return db
}

func initFakeDataBase(databasePath string) *gorm.DB {
	dbLogger := logger.Default
	if viper.GetBool(enableDbInfoLogs) {
		dbLogger = dbLogger.LogMode(logger.Info)
	}

	temp, _ := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		Logger: dbLogger,
	})
	// this will ensure table is created
	if err := temp.AutoMigrate(&APIEvent{}, &APIInfo{}, &Review{}, &APIPath{}); err != nil {
		panic(err)
	}

	return temp
}
