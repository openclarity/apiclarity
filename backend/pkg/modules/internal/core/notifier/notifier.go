package notifier

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/openclarity/apiclarity/api3/notifications"
)

type notificationWithParams struct {
	apiID uint
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
