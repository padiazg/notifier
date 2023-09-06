package engine

import (
	"sync"

	"github.com/google/uuid"
	n "github.com/padiazg/notifier/notification"
)

// NotificationEngine handles the dispatch and tracking of notifications
type NotificationEngine struct {
	Webhook n.Notifier
	MQ      n.Notifier
	OnError func(error)
}

func (ne *NotificationEngine) Dispatch(notification *n.Notification) {
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
