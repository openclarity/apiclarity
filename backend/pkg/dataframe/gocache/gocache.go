// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

type DataFrameGoCache struct {
	backend *cache.Cache
}

func NewDataFrame() (*DataFrameGoCache, error) {
	df := DataFrameGoCache{}
	err := df.Init()
	if err != nil {
		return nil, err
	}

	return &df, nil
}

func (df *DataFrameGoCache) Init() error {
	df.backend = cache.New(5*time.Minute, 10*time.Minute)

	return nil
}

func (df *DataFrameGoCache) Set(key string, value interface{}, ttl time.Duration) bool {
	df.backend.Set(key, value, ttl)

	return true
}

func (df *DataFrameGoCache) Get(key string) (interface{}, bool) {
	return df.backend.Get(key)
}

func (df *DataFrameGoCache) Del(key string) {
	df.backend.Delete(key)
}
