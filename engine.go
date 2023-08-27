package notifier

import (
	"fmt"
	"sync"
)

// NotificationEngine handles the dispatch and tracking of notifications
type NotificationEngine struct {
	Webhook Notifier
	MQ      Notifier
	// You can add other fields for tracking, logging, etc.
}

func (ne *NotificationEngine) DispatchAndTrack(event EventType, data interface{}) {
	notification := Notification{Event: event, Data: data}

	// Dispatch asynchronously using goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		ne.Webhook.SendNotification(notification)
	}()

	go func() {
		defer wg.Done()
		ne.MQ.SendNotification(notification)
	}()

	wg.Wait()

	// Save notification and result to a table for tracking
	ne.saveNotificationToTable(notification)
}

func (ne *NotificationEngine) saveNotificationToTable(notification Notification) {
	// Implement saving logic here
	fmt.Printf("Saved notification to table: %+v\n", notification)
}
