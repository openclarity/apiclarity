// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
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

// nolint: revive,stylecheck
package spec_differ

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api/server/models"
	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/database"
)

func (s *specDiffer) StartDiffsSender(ctx context.Context) {
	// each period aggregate diffs per api and notify to notification server
	log.Info("Starting diffs sender")
	interval := s.config.SendNotificationIntervalSec()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.sendDiffsNotifications(); err != nil {
				log.Errorf("Failed to send diffs notification. total diffs=%v.: %v", s.totalUniqueDiffs, err)
			}
			s.clearDiffs()
		}
	}
}

func (s *specDiffer) clearDiffs() {
	s.Lock()
	defer s.Unlock()
	s.apiIDToDiffs = map[uint]map[diffHash]global.Diff{}
	s.totalUniqueDiffs = 0
}

func (s *specDiffer) sendDiffsNotifications() error {
	if s.getTotalUniqueDiffs() == 0 {
		log.Infof("No events to send")
		return nil
	}

	diffsNotifications := s.getSpecDiffsNotifications()

	log.Infof("Sending diff notifications: %+v", diffsNotifications)

	for _, notification := range diffsNotifications {
		n := notifications.APIClarityNotification{}
		if err := n.FromSpecDiffsNotification(notification); err != nil {
			return fmt.Errorf("failed to convert to apiclarity notification: %v", err)
		}
		apiID := *notification.Diffs.ApiInfo.Id
		if err := s.accessor.Notify(context.TODO(), moduleName, uint(apiID), n); err != nil {
			return fmt.Errorf("failed to notify: %v", err)
		}
	}

	return nil
}

func (s *specDiffer) getTotalUniqueDiffs() int {
	s.RLock()
	defer s.RUnlock()
	return s.totalUniqueDiffs
}

func (s *specDiffer) getSpecDiffsNotifications() []notifications.SpecDiffsNotification {
	s.RLock()
	defer s.RUnlock()

	var ret []notifications.SpecDiffsNotification

	for apiID, apiInfoDiffs := range s.apiIDToDiffs {
		apiInfo, err := s.accessor.GetAPIInfo(context.TODO(), apiID)
		if err != nil {
			log.Errorf("Failed to get api info with apiID=%v: %v", apiID, err)
			continue
		}
		var diffs []global.Diff
		for _, diff := range apiInfoDiffs {
			diffs = append(diffs, diff)
		}
		ret = append(ret, notifications.SpecDiffsNotification{
			Diffs: global.APIDiffs{
				ApiInfo: convertAPIInfo(apiInfo),
				Diffs:   diffs,
			},
		})
	}

	return ret
}

func convertAPIInfo(apiInfo *database.APIInfo) common.ApiInfoWithType {
	id := uint32(apiInfo.ID)
	port := int(apiInfo.Port)
	return common.ApiInfoWithType{
		ApiType:              convertAPIType(apiInfo.Type),
		DestinationNamespace: &apiInfo.DestinationNamespace,
		HasProvidedSpec:      &apiInfo.HasProvidedSpec,
		HasReconstructedSpec: &apiInfo.HasReconstructedSpec,
		Id:                   &id,
		Name:                 &apiInfo.Name,
		Port:                 &port,
		TraceSourceId:        &apiInfo.TraceSource.UID,
	}
}

func convertAPIType(apiType models.APIType) *common.ApiTypeEnum {
	switch apiType {
	case models.APITypeINTERNAL:
		typ := common.INTERNAL
		return &typ
	case models.APITypeEXTERNAL:
		typ := common.EXTERNAL
		return &typ
	default:
		log.Errorf("Unknown api type: %v", apiType)
		typ := common.INTERNAL
		return &typ
	}
}
