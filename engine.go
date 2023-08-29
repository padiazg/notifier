package notifier

import (
	"sync"

	"github.com/google/uuid"
)

// NotificationEngine handles the dispatch and tracking of notifications
type NotificationEngine struct {
	Webhook Notifier
	MQ      Notifier
	OnError func(error)
}

func (ne *NotificationEngine) Dispatch(notification *Notification) {
	if notification == nil {
		return
	}

	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	var wg sync.WaitGroup

	if ne.Webhook != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := ne.Webhook.SendNotification(notification)
			if !res.Success && ne.OnError != nil {
				ne.OnError(res.Error)
			}
		}()
	}

	if ne.MQ != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := ne.MQ.SendNotification(notification)
			if !res.Success && ne.OnError != nil {
				ne.OnError(res.Error)
			}
		}()
	}

	wg.Wait()
}
