// Copyright © 2023 Cisco Systems, Inc. and its affiliates.
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

package lineage

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

type ParentMap struct {
	cache  *ristretto.Cache
	keyTTL time.Duration
}

func NewParentMap(size int64, ttl time.Duration) (*ParentMap, error) {
	config := ristretto.Config{
		NumCounters:        (size / 8) * 10,
		MaxCost:            size,
		BufferItems:        64,
		Metrics:            false,
		IgnoreInternalCost: false,
	}
	if cache, err := ristretto.NewCache(&config); err != nil {
		return nil, err
	} else {
		return &ParentMap{
			cache:  cache,
			keyTTL: ttl, //time.Duration in nanoseconds
		}, nil
	}
}

func (m *ParentMap) GetParent(id string) (string, bool) {
	//var key []byte = id[:]
	parent, found := m.cache.Get(id)
	if !found {
		return "", false
	}
	switch d := parent.(type) {
	default:
		return "", false
	case string:
		return d, true
	}
}

func (m *ParentMap) PutParent(id string, parent string) bool {
	if id == "" || parent == "" {
		return false
	}
	//var key []byte = id[:]
	success := m.cache.SetWithTTL(id, parent, int64(len(parent)), m.keyTTL)
	if success {
		m.cache.Wait() //wait on insert, not lookup
	}
	return success
}
