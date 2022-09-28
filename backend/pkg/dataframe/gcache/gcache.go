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

package gcache

import (
	"time"

	"github.com/bluele/gcache"
)

type DataFrameGCache struct {
	backend gcache.Cache
}

func NewDataFrame() (*DataFrameGCache, error) {
	df := DataFrameGCache{}
	err := df.Init()
	if err != nil {
		return nil, err
	}

	return &df, nil
}

func (df *DataFrameGCache) Init() error {
	backend := gcache.New(20).
		LRU().
		Build()
	df.backend = backend

	return nil
}

func (df *DataFrameGCache) Set(key string, value interface{}, ttl time.Duration) bool {
	if ttl == 0 {
		df.backend.Set(key, value)
	} else {
		df.backend.SetWithExpire(key, value, time.Second*10)
	}

	return true
}

func (df *DataFrameGCache) Get(key string) (interface{}, bool) {
	result, err := df.backend.Get(key)
	if err != nil {
		return nil, false
	}

	return result, true
}

func (df *DataFrameGCache) Del(key string) {
	df.backend.Remove(key)
}
