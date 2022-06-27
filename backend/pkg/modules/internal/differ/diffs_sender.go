package differ

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api3/common"
	"github.com/openclarity/apiclarity/api3/global"
	"github.com/openclarity/apiclarity/api3/notifications"
	"github.com/openclarity/apiclarity/backend/pkg/database"
)

func (p *differ) StartDiffsSender(ctx context.Context) {
	// each period aggregate diffs per api and notify to notification server
	log.Info("Starting diffs sender")
	interval := p.config.SendNotificationIntervalSec()

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.sendDiffsNotifications(); err != nil {
				log.Errorf("Failed to send diffs notification. total diffs=%v.: %v", p.totalDiffEvents, err)
			}
			p.clearDiffs()
		}
	}
}

func (p *differ) clearDiffs() {
	p.Lock()
	defer p.Unlock()
	p.apiIDToDiffs = map[uint]map[diffHash]global.Diff{}
	p.totalDiffEvents = 0
}

func (p *differ) sendDiffsNotifications() error {
	if p.getTotalDiffEvents() == 0 {
		log.Infof("No events to send")
		return nil
	}

	diffsNotifications := p.getSpecDiffsNotifications()

	log.Infof("Sending diff notifications: %+v", diffsNotifications)

	for _, notification := range diffsNotifications {
		n := notifications.APIClarityNotification{}
		if err := n.FromSpecDiffsNotification(notification); err != nil {
			return fmt.Errorf("failed to convert to apiclarity notification: %v", err)
		}
		apiID := *notification.Diffs.ApiInfo.Id
		if err := p.accessor.Notify(context.TODO(), moduleName, uint(apiID), n); err != nil {
			return fmt.Errorf("failed to notify: %v", err)
		}
	}

	return nil
}

func (p *differ) getTotalDiffEvents() int {
	p.RLock()
	defer p.RUnlock()
	return p.totalDiffEvents
}

func (p *differ) getSpecDiffsNotifications() []notifications.SpecDiffsNotification {
	p.RLock()
	defer p.RUnlock()

	var ret []notifications.SpecDiffsNotification

	for apiID, apiInfoDiffs := range p.apiIDToDiffs {
		apiInfo, err := p.accessor.GetAPIInfo(context.TODO(), apiID)
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

func convertAPIInfo(apiInfo *database.APIInfo) common.ApiInfo {
	id := uint32(apiInfo.ID)
	port := int(apiInfo.Port)
	return common.ApiInfo{
		DestinationNamespace: &apiInfo.DestinationNamespace,
		HasProvidedSpec:      &apiInfo.HasProvidedSpec,
		HasReconstructedSpec: &apiInfo.HasReconstructedSpec,
		Id:                   &id,
		Name:                 &apiInfo.Name,
		Port:                 &port,
	}
}
