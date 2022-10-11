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

package gocache

import (
	"time"

	cache "github.com/patrickmn/go-cache"
)

const defaultCleanupInterval = 5 * time.Minute

type DataFrame struct {
	backend *cache.Cache
}

func (df *DataFrame) Init(ttl time.Duration) error {
	df.backend = cache.New(ttl, defaultCleanupInterval)

	return nil
}

func (df *DataFrame) Set(key string, value interface{}) bool {
	df.backend.Set(key, value, 0)

	return true
}

func (df *DataFrame) Get(key string) (interface{}, bool) {
	return df.backend.Get(key)
}

func (df *DataFrame) Del(key string) {
	df.backend.Delete(key)
}
