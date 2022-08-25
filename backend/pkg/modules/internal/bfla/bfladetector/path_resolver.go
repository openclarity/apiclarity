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
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/backend/pkg/database"
)

func ParseSpecInfo(apiInfo *database.APIInfo) ([]*models.SpecTag, error) {
	if apiInfo.ProvidedSpecInfo != "" {
		info := &models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ProvidedSpecInfo), info); err != nil {
			return nil, fmt.Errorf("unable to unmarshal spec tags: %s", err)
		}
		return info.Tags, nil
	}
	if apiInfo.ReconstructedSpecInfo != "" {
		info := &models.SpecInfo{}
		if err := json.Unmarshal([]byte(apiInfo.ReconstructedSpecInfo), info); err != nil {
			log.Errorf("unable to unmarshal spec tags: %s", err)
			return nil, fmt.Errorf("unable to unmarshal spec tags: %s", err)
		}
		return info.Tags, nil
	}
	return nil, nil
}

func ResolvePath(tags []*models.SpecTag, event *database.APIEvent) (string, error) {
	if event.ProvidedPathID != "" {
		return resolvePathFromTags(tags, event.ProvidedPathID), nil
	}
	if event.ReconstructedPathID != "" {
		return resolvePathFromTags(tags, event.ReconstructedPathID), nil
	}
	return "", fmt.Errorf("Event %v cannot resolve to a spec path", event.ID)
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

func resolveTagsForPathAndMethod(tags []*models.SpecTag, path, method string) (tagNames []string) {
	for _, tag := range tags {
		for _, methodAndPath := range tag.MethodAndPathList {
			if path == methodAndPath.Path && string(methodAndPath.Method) == method {
				tagNames = append(tagNames, tag.Name)
			}
		}
	}
	return tagNames
}
