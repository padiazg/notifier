package notifier

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQP09Notifier implements the Notifier interface for message queues
type AMQP09Notifier struct {
	QueueName string // The name of the message queue
	conn      *amqp.Connection
	channel   *amqp.Channel
}

// SendNotification sends a notification to the message queue
func (mn *AMQP09Notifier) SendNotification(notification *Notification) NotificationResult {
	// Serialize the notification data to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}

	// Connect to the message queue and send the payload
	err = mn.send(payload)
	if err != nil {
		return NotificationResult{Success: false, Error: err}
	}

	return NotificationResult{Success: true}
}

// send sends a message to the message queue
func (mn *AMQP09Notifier) send(payload []byte) error {
	// Simulate connecting to the message queue and sending the payload
	// Replace this with actual message queue library usage
	fmt.Printf("Sending notification to queue '%s': %s\n", mn.QueueName, payload)
	return nil
}

// Connect connects to the message queue
// this can used to reconnect to the message queue in case of a failure
func (mn *AMQP09Notifier) Connect() error {

	return nil
}

// Close closes the connection to the message queue
func (mn *AMQP09Notifier) Close() error {
	return mn.conn.Close()
}
