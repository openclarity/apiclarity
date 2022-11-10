// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package config

import (
	"sync"

	"github.com/spf13/viper"
)

const (
	sendNotificationIntervalSec = "DIFFER_SEND_NOTIFICATION_INTERVAL_SEC"
	diffsSendThreshold          = "DIFF_SEND_THRESHOLD"
)

type Config struct {
	sendNotificationIntervalSec int
	diffsSendThreshold          int
}

// the Mutex to restrict access to configSingleton.
var lock sync.Mutex

// config singleton.
var configSingleton *Config

// return the singleton instance.
func GetConfig() *Config {
	lock.Lock()
	defer lock.Unlock()

	if configSingleton == nil {
		configSingleton = NewDifferConfig()
	}
	return configSingleton
}

func (c *Config) SendNotificationIntervalSec() int {
	return c.sendNotificationIntervalSec
}

func (c *Config) DiffsSendThreshold() int {
	return c.diffsSendThreshold
}

func NewDifferConfig() *Config {
	viper.SetDefault(sendNotificationIntervalSec, "300")
	viper.SetDefault(diffsSendThreshold, "500")

	config := Config{
		sendNotificationIntervalSec: viper.GetInt(sendNotificationIntervalSec),
		diffsSendThreshold:          viper.GetInt(diffsSendThreshold),
	}

	return &config
}
