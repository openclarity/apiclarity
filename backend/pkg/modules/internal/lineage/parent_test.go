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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultMapSize = 8192
	defaultMapTTL  = time.Second * 60
)

func TestParentInit(t *testing.T) {
	_, err := NewTraceParentMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
}

func TestParentGetEmpty(t *testing.T) {
	parentMap, err := NewTraceParentMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
	spanID := generateSpanID()
	parent, found := parentMap.GetParent(spanID)
	assert.False(t, found)
	assert.Equal(t, "", parent)
}

func TestParentPuEmpty(t *testing.T) {
	parentMap, err := NewTraceParentMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
	spanID := generateSpanID()
	success := parentMap.PutParent(spanID, "")
	assert.False(t, success)
}

func TestParentPut(t *testing.T) {
	parentMap, err := NewTraceParentMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
	spanID := generateSpanID()
	insertParent := "parent1"
	success := parentMap.PutParent(spanID, insertParent)
	assert.True(t, success)
	parent, found := parentMap.GetParent(spanID)
	assert.True(t, found)
	assert.Equal(t, insertParent, parent)
}
