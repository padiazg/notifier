package amqp

import (
	"fmt"

	"github.com/padiazg/notifier/notification"
	n "github.com/padiazg/notifier/notification"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQP09Notifier implements the Notifier interface for message queues
type AMQP09Notifier struct {
	*Config
	conn    *amqp.Connection
	channel *amqp.Channel
	Channel chan *n.Notification
}

func (n *AMQP09Notifier) Type() string {
	return "amqp09"
}

func (n *AMQP09Notifier) Name() string {
	return n.Config.Name
}

func (n *AMQP09Notifier) StartWorker() {
	fmt.Println("AMQP10Notifier.StartWorker")
	n.Channel = make(chan *notification.Notification)
	for notification := range n.Channel {
		n.SendNotification(notification)
	}

	fmt.Printf("AMQP10 notifier stopped\n")
}

func (n *AMQP09Notifier) GetChannel() chan *notification.Notification {
	return n.Channel
}

func (n *AMQP09Notifier) Notify(payload *notification.Notification) {
	n.Channel <- payload
}

func (n *AMQP09Notifier) SendNotification(payload *n.Notification) notification.Result {
	return notification.Result{Success: true}
}
