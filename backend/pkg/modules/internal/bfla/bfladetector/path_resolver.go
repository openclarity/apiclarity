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

package bfladetector

import (
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/database"
)

func ResolvePath(apiInfo *database.APIInfo, event *database.APIEvent) (urlpath string) {
	if event.ProvidedPathID != "" && apiInfo.ProvidedSpecInfo != "" {
		info := &models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ProvidedSpecInfo), info); err != nil {
			log.Errorf("unable to unmarshal spec tags: %s", err)
			return ""
		}
		return resolvePathFromTags(info.Tags, event.ProvidedPathID)
	}
	if event.ReconstructedPathID != "" && apiInfo.ReconstructedSpecInfo != "" {
		info := &models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ReconstructedSpecInfo), info); err != nil {
			log.Errorf("unable to unmarshal spec tags: %s", err)
			return ""
		}
		return resolvePathFromTags(info.Tags, event.ReconstructedPathID)
	}
	return ""
}

func resolvePathFromTags(tags []*models.SpecTag, pathID string) string {
	for _, tag := range tags {
		for _, methodAndPath := range tag.MethodAndPathList {
			if pathID == string(methodAndPath.PathID) {
				return methodAndPath.Path
			}
		}
	}
	return ""
}
