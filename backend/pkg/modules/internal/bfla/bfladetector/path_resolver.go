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

func ResolvePath(tags []*models.SpecTag, event *database.APIEvent) (path string, tagNames []string, err error) {
	if event.ProvidedPathID != "" {
		path, tagNames, err = resolvePathFromPathIDAndMethod(tags, event.ProvidedPathID, string(event.Method))
	} else if event.ReconstructedPathID != "" {
		path, tagNames, err = resolvePathFromPathIDAndMethod(tags, event.ReconstructedPathID, string(event.Method))
	} else {
		err = fmt.Errorf("event %v cannot resolve to a spec path", event.ID)
	}
	return path, tagNames, err
}

func resolvePathFromPathIDAndMethod(tags []*models.SpecTag, pathID string, method string) (path string, tagNames []string, err error) {
	for _, tag := range tags {
		for _, methodAndPath := range tag.MethodAndPathList {
			if pathID == string(methodAndPath.PathID) && string(methodAndPath.Method) == method {
				path = methodAndPath.Path
				tagNames = append(tagNames, tag.Name)
			}
		}
	}
	if path == "" {
		err = fmt.Errorf("unable to resolve pathId %v  / method %v", pathID, method)
	}
	return path, tagNames, err
}

func resolveTagsFromPathAndMethod(tags []*models.SpecTag, path, method string) (tagNames []string, err error) {
	for _, tag := range tags {
		for _, methodAndPath := range tag.MethodAndPathList {
			if path == methodAndPath.Path && string(methodAndPath.Method) == method {
				tagNames = append(tagNames, tag.Name)
			}
		}
	}
	if len(tagNames) == 0 {
		err = fmt.Errorf("unable to resolve path %v / method %v", path, method)
	}

	return tagNames, err
}
