package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/Azure/go-amqp"
	n "github.com/padiazg/notifier/notification"
)

// AMQP10Notifier implements the Notifier interface for message queues
type AMQP10Notifier struct {
	*Config
	conn    *amqp.Conn
	session *amqp.Session
	sender  *amqp.Sender
	ctx     context.Context
}

func NewAMQP10Notifier(config *Config) *AMQP10Notifier {
	return &AMQP10Notifier{
		Config: config,
	}
}

// SendNotification sends a notification to the message queue
func (mn *AMQP10Notifier) SendNotification(notification *n.Notification) n.NotificationResult {
	// Serialize the notification data to JSON
	payload, err := json.Marshal(notification)
	if err != nil {
		return n.NotificationResult{Success: false, Error: err}
	}

	// Connect to the message queue and send the payload
	err = mn.send(payload)
	if err != nil {
		return n.NotificationResult{Success: false, Error: err}
	}

	return n.NotificationResult{Success: true}
}

// send sends a message to the message queue
// https://pkg.go.dev/github.com/Azure/go-amqp#example-package
func (mn *AMQP10Notifier) send(payload []byte) error {
	var (
		ctx, cancel = context.WithTimeout(mn.ctx, 3*time.Second)
		err         error
	)
	defer cancel()

	// send message
	err = mn.sender.Send(ctx, amqp.NewMessage(payload), nil)
	if err != nil {
		return fmt.Errorf("sending message: %v", err)
	}

	return nil
}

// Connect connects to the message queue
// this can used to reconnect to the message queue in case of a failure
func (mn *AMQP10Notifier) Connect() error {
	var err error

	// create a context
	mn.ctx = context.TODO()

	// create a connection
	mn.conn, err = amqp.Dial(mn.ctx, mn.Address, nil)
	if err != nil {
		return fmt.Errorf("dialing AMQP server: %v", err)
	}

	// create a session
	mn.session, err = mn.conn.NewSession(mn.ctx, nil)
	if err != nil {
		return fmt.Errorf("creating AMQP session: %v", err)
	}

	// create a sender
	mn.sender, err = mn.session.NewSender(mn.ctx, mn.QueueName, &amqp.SenderOptions{
		TargetDurability: amqp.DurabilityUnsettledState,
	})
	if err != nil {
		return fmt.Errorf("creating sender link: %v", err)
	}

	return nil
}

// Close closes the connection to the message queue
func (mn *AMQP10Notifier) Close() error {
	return mn.conn.Close()
}
