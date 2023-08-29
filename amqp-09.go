package notifier

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MQNotifier implements the Notifier interface for message queues
type AMQP09Notifier struct {
	QueueName string // The name of the message queue
	conn      *amqp.Connection
	ch        *amqp.Channel
}

func (mn *AMQP09Notifier) SendNotification(notification *Notification) NotificationResult {
	// Serialize the notification data to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}

	// Connect to the message queue and send the payload
	err = mn.connectAndSend(payload)
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}

	return NotificationResult{Success: true}
}

func (mn *AMQP09Notifier) connectAndSend(payload []byte) error {
	// Simulate connecting to the message queue and sending the payload
	// Replace this with actual message queue library usage
	fmt.Printf("Sending notification to queue '%s': %s\n", mn.QueueName, payload)
	return nil
}
