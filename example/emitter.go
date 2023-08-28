package main

import "github.com/padiazg/notifier"

const (
	UserCreated notifier.EventType = "UserCreated"
	UserUpdated notifier.EventType = "UserUpdated"
	UserDeleted notifier.EventType = "UserDeleted"
)

func main() {
	// Initialize notifiers
	webhookNotifier := &notifier.WebhookNotifier{Endpoint: "http://localhost:8080/webhook"}
	mqNotifier := &notifier.AMQP10Notifier{}

	// Initialize the notification engine
	notificationEngine := &notifier.NotificationEngine{
		Webhook: webhookNotifier,
		MQ:      mqNotifier,
	}

	// Example usage
	event := UserCreated
	data := "sample data"
	notificationEngine.DispatchAndTrack(event, data)
}
