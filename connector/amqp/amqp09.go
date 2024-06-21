package amqp

import (
	"fmt"

	"github.com/padiazg/notifier/notification"
	"github.com/padiazg/notifier/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQP09Notifier implements the Notifier interface for message queues
type AMQP09Notifier struct {
	*Config
	conn *amqp.Connection
	// channel *amqp.Channel
	Channel chan *notification.Notification
}

func (n *AMQP09Notifier) New(config *Config) *AMQP09Notifier {
	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = n.Type() + utils.RandomId8()
	}

	config.Protocol = ProtocolAMQP09

	n.Config = config

	return n
}

func (n *AMQP09Notifier) Type() string {
	return "amqp09"
}

func (n *AMQP09Notifier) Name() string {
	return n.Config.Name
}

func (n *AMQP09Notifier) GetChannel() chan *notification.Notification {
	return n.Channel
}

func (n *AMQP09Notifier) Connect() error {
	var err error

	n.conn, err = amqp.Dial(n.Address)
	if err != nil {
		return fmt.Errorf("dialing AMQP server: %w", err)
	}

	return nil
}

func (n *AMQP09Notifier) StartWorker() {
	n.Channel = make(chan *notification.Notification)
	for notification := range n.Channel {
		n.SendNotification(notification)
	}
}

func (n *AMQP09Notifier) Notify(payload *notification.Notification) {
	n.Channel <- payload
}

func (n *AMQP09Notifier) SendNotification(payload *notification.Notification) *notification.Result {
	// TODO: implement!
	return &notification.Result{Success: false}
}

func (n *AMQP09Notifier) Close() error {
	return n.conn.Close()
}
