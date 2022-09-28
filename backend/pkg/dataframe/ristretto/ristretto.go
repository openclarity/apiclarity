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

package ristretto

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
)

type DataFrame struct {
	backend *ristretto.Cache
	ttl     time.Duration
}

func (df *DataFrame) Init(ttl time.Duration) error {
	df.ttl = ttl
	config := &ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	}
	backend, err := ristretto.NewCache(config)
	if err != nil {
		return fmt.Errorf("unable to initialize dataframe: %v", err)
	}
	df.backend = backend

	return nil
}

func (df *DataFrame) Set(key string, value interface{}) bool {
	cost := int64(0)
	return df.backend.SetWithTTL(key, value, cost, df.ttl)
}

func (df *DataFrame) Get(key string) (interface{}, bool) {
	return df.backend.Get(key)
}

func (df *DataFrame) Del(key string) {
	df.backend.Del(key)
}
