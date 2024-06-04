package amqp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/padiazg/notifier/notification"
	"github.com/padiazg/notifier/utils"
)

// AMQP10Notifier implements the Notifier interface for message queues
type AMQP10Notifier struct {
	*Config
	conn    *amqp.Conn
	session *amqp.Session
	sender  *amqp.Sender
	ctx     context.Context
	Channel chan *notification.Notification
}

func (n *AMQP10Notifier) New(config *Config) *AMQP10Notifier {
	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = n.Type() + utils.RamdomId8()
	}

	config.Protocol = ProtocolAMQP10

	n.Config = config

	return n
}

func (n *AMQP10Notifier) Type() string {
	return "amqp10"
}

func (n *AMQP10Notifier) Name() string {
	return n.Config.Name
}

func (n *AMQP10Notifier) GetChannel() chan *notification.Notification {
	return n.Channel
}

func (n *AMQP10Notifier) Connect() error {
	// fmt.Println("AMQP10Notifier.Connect")
	var err error

	// create a context
	n.ctx = context.TODO()

	// create a connection
	// fmt.Println("AMQP10Notifier.StartWorker Dial")
	n.conn, err = amqp.Dial(n.ctx, n.Address, nil)
	if err != nil {
		return fmt.Errorf("dialing AMQP server: %v", err)
	}

	// create a session
	// fmt.Println("AMQP10Notifier.StartWorker NewSession")
	n.session, err = n.conn.NewSession(n.ctx, nil)
	if err != nil {
		return fmt.Errorf("creating AMQP session: %v", err)
	}

	// create a sender
	// fmt.Println("AMQP10Notifier.StartWorker NewSender")
	n.sender, err = n.session.NewSender(n.ctx, n.QueueName, &amqp.SenderOptions{
		TargetDurability: amqp.DurabilityUnsettledState,
	})
	if err != nil {
		return fmt.Errorf("creating sender link: %v", err)
	}

	return nil
}

func (n *AMQP10Notifier) Close() error {
	// fmt.Println("AMQP10Notifier.Close")
	return n.conn.Close()
}

func (n *AMQP10Notifier) Notify(payload *notification.Notification) {
	// fmt.Printf("AMQP10Notifier.Notify: %v channel:%v\n", payload.ID, n.Channel)
	if n.Channel == nil || payload == nil {
		return
	}
	n.Channel <- payload
	// fmt.Printf("AMQP10Notifier.Notify: sent\n")
}

func (n *AMQP10Notifier) StartWorker() {
	// fmt.Println("AMQP10Notifier.StartWorker")

	// create a channel
	n.Channel = make(chan *notification.Notification)

	for notification := range n.Channel {
		res := n.SendNotification(notification)
		if !res.Success {
			fmt.Printf("AMQP Success: %v\n", res.Success) // TODO: implement logger or pass to error handler
		}
	}

	// close the sender
	if err := n.Close(); err != nil {
		fmt.Printf("closing sender link: %v", err) // TODO: implement logger or pass to error handler
	}

	// fmt.Printf("AMQP10 notifier stopped\n")
}

func (n *AMQP10Notifier) SendNotification(message *notification.Notification) *notification.Result {
	// fmt.Println("AMQP10Notifier.SendNotification")
	var (
		ctx, cancel = context.WithTimeout(n.ctx, 3*time.Second)
		err         error
	)
	defer cancel()

	// Serialize the notification data to JSON
	payload, err := json.Marshal(message)
	if err != nil {
		return &notification.Result{Success: false, Error: err}
	}

	// send message
	err = n.sender.Send(ctx, amqp.NewMessage(payload), nil)
	if err != nil {
		return &notification.Result{Success: false, Error: fmt.Errorf("sending message: %v", err)}
	}

	return &notification.Result{Success: true}
}
