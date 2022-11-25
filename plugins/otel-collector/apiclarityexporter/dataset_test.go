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

package apiclarityexporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultMapSize = 1024
	defaultMapTTL  = time.Second * 10
)

func TestDatasetInit(t *testing.T) {
	_, err := NewTraceDatasetMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
}

func TestDatasetGetEmpty(t *testing.T) {
	datasetMap, err := NewTraceDatasetMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
	spanID := generateSpanID()
	dataset, found := datasetMap.GetDataset(spanID)
	assert.False(t, found)
	assert.Equal(t, "", dataset)
}

func TestDatasetPuEmpty(t *testing.T) {
	datasetMap, err := NewTraceDatasetMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
	spanID := generateSpanID()
	success := datasetMap.PutDataset(spanID, "")
	assert.False(t, success)
}

func TestDatasetPut(t *testing.T) {
	datasetMap, err := NewTraceDatasetMap(defaultMapSize, defaultMapTTL)
	require.NoError(t, err)
	spanID := generateSpanID()
	insertDataset := "dataset1"
	success := datasetMap.PutDataset(spanID, insertDataset)
	assert.True(t, success)
	dataset, found := datasetMap.GetDataset(spanID)
	assert.True(t, found)
	assert.Equal(t, insertDataset, dataset)
}
