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

package dataframe

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/openclarity/apiclarity/backend/pkg/dataframe/gcache"
	"github.com/openclarity/apiclarity/backend/pkg/dataframe/ristretto"
	"github.com/openclarity/apiclarity/backend/pkg/dataframe/gocache"
)

func UnknownKey(t *testing.T, df DataFrame) {
	_, found := df.Get("unknown")
	if found {
		t.Fatalf("Key 'unknown' shouldn't be present")
	}
}

func Set(t *testing.T, df DataFrame) {
	isSet := df.Set("key1", "value1", 0)
	if !isSet {
		t.Fatalf("Error while setting value")
	}
}

func SetGet(t *testing.T, df DataFrame) {
	df.Set("key1", "value1", 10 * time.Minute)
	time.Sleep(1 * time.Millisecond) // Let time for the admission policy
	result, found := df.Get("key1")
	if !found {
		t.Fatalf("Key 'key1' was not found")
	}
	if result != "value1" {
		t.Fatalf("Value 'value1' was expected but got '%s'", result)
	}
}

func Del(t *testing.T, df DataFrame) {
	df.Set("key1", "value1", 10 * time.Minute)
	df.Set("key2", "value2", 10 * time.Minute)
	time.Sleep(1 * time.Millisecond)
	df.Del("key1")
	_, found := df.Get("key1")
	if found {
		t.Fatalf("Key 'key1' must be absent because it was deleted")
	}
	_, found = df.Get("key2")
	if !found {
		t.Fatalf("Key 'key2' must still be present")
	}
}

func TestBackends(t *testing.T) {
	ristrettoCache, err := ristretto.NewDataFrame()
	if err != nil {
		t.Fatalf("Unable to initialize ristretto cache backend")
	}
	gcacheCache, err := gcache.NewDataFrame()
	if err != nil {
		t.Fatalf("Unable to initialize gcache cache backend")
	}
	gocacheCache, err := gocache.NewDataFrame()
	if err != nil {
		t.Fatalf("Unable to initialize gocache cache backend")
	}

	for _, b := range []DataFrame{ristrettoCache, gcacheCache, gocacheCache} {
		t.Run(fmt.Sprintf("Backend %s", reflect.TypeOf(b)), func(t *testing.T) {
			UnknownKey(t, b)
			Set(t, b)
			SetGet(t, b)
			Del(t, b)
		})
	}
}
