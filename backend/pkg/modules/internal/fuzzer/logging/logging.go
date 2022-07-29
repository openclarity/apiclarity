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

package logging

import (
	log "github.com/sirupsen/logrus"
)

// Logf logs message either via defined user logger or via system one if no user logger is defined.

func Debugf(f string, args ...interface{}) {
	log.Debugf(f, args...)
}

func Logf(f string, args ...interface{}) {
	log.Infof(f, args...)
}

func Warningf(f string, args ...interface{}) {
	log.Warningf(f, args...)
}

func Errorf(f string, args ...interface{}) {
	log.Errorf(f, args...)
}

func InitLogger() {
	// Use this function to init a logger if don't use the root one
}
