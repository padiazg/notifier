package engine

import (
	"sync"

	"github.com/google/uuid"
	ad "github.com/padiazg/notifier/drivers/amqp"
	wd "github.com/padiazg/notifier/drivers/webhook"
	n "github.com/padiazg/notifier/notification"
)

// NotificationEngine handles the dispatch and tracking of notifications
type NotificationEngine struct {
	WebHook n.Notifier
	MQ      n.Notifier
	OnError func(error)
}

func NewNotificationEngine(config *Config) *NotificationEngine {
	notificationEngine := &NotificationEngine{}

	if config == nil {
		return notificationEngine
	}

	if config.WebHook != nil {
		notificationEngine.WebHook = wd.NewWebhookNotifier(config.WebHook)
	}

	if config.MQ != nil {
		notificationEngine.MQ = ad.NewAMQPNotifier(config.MQ)
	}

	if config.OnError != nil {
		notificationEngine.OnError = config.OnError
	}

	return notificationEngine
}

func (ne *NotificationEngine) Dispatch(notification *n.Notification) {
	if notification == nil {
		return
	}

	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	var wg sync.WaitGroup

	if ne.WebHook != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := ne.WebHook.SendNotification(notification)
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
