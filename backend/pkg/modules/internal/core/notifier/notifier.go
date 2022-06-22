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

package notifier

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api3/notifications"
)

type notificationWithParams struct {
	apiID   uint
	payload notifications.APIClarityNotification
}

type Notifier struct {
	notificationURL   string
	notificationQueue chan notificationWithParams
	workers           int
}

func NewNotifier(notificationPrefixURL string, maxQueueSize int, workers int) *Notifier {
	return &Notifier{
		notificationURL:   notificationPrefixURL,
		notificationQueue: make(chan notificationWithParams, maxQueueSize),
		workers:           workers,
	}
}

func (n Notifier) Start(ctx context.Context) {
	for i := 0; i < n.workers; i++ {
		go worker(ctx, n.notificationURL, n.notificationQueue)
	}
}

func (n Notifier) Stop() {
	close(n.notificationQueue)
}

func (n Notifier) Notify(apiID uint, notif notifications.APIClarityNotification) error {
	n.notificationQueue <- notificationWithParams{apiID: apiID, payload: notif}

	return nil
}

func worker(ctx context.Context, notificationPrefixURL string, notifQueue <-chan notificationWithParams) {
	c, err := notifications.NewClient(notificationPrefixURL)
	if err != nil {
		log.Errorf("unable to create notification client: %s", err)
		return
	}

	for {
		select {
		case notification, ok := <-notifQueue:
			if !ok {
				return
			}
			log.Debugf("[CORE] Notification in progress to apiID=%d...", notification.apiID)
			resp, err := c.PostNotificationApiID(ctx, int64(notification.apiID), notification.payload)
			if err != nil {
				log.Errorf("error while sending notification to '%s': %s", notificationPrefixURL, err)
				continue
			}
			resp.Body.Close()
		case <-ctx.Done():
			return
		}
	}
}
