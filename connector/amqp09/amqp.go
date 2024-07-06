package amqp09

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/padiazg/notifier/notification"
	"github.com/padiazg/notifier/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	// amqp "github.com/rabbitmq/amqp091-go"
)

type PublishOptions struct {
	Key       string
	Mandatory bool
	Immediate bool
}

type Config struct {
	Logger          *log.Logger
	Name            string
	QueueName       string
	Address         string
	DeliveryTimeout time.Duration
	PublishOptions  PublishOptions
	wrapper         internalWrapperInterface
	ctx             context.Context
}

// AMQPNotifier implements the Notifier interface for message queues
type AMQPNotifier struct {
	*Config
	Channel     chan *notification.Notification
	jsonMarshal func(v any) ([]byte, error)
}

var _ notification.Notifier = (*AMQPNotifier)(nil)

func New(config *Config) *AMQPNotifier {
	return (&AMQPNotifier{}).New(config)
}

func (n *AMQPNotifier) New(config *Config) *AMQPNotifier {
	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = n.Type() + utils.RandomId8()
	}

	if config.Logger == nil {
		config.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	if config.ctx == nil {
		config.ctx = context.TODO()
	}

	if config.wrapper == nil {
		config.wrapper = &internalWrapper{}
	}

	n.Config = config
	n.jsonMarshal = json.Marshal
	n.Channel = make(chan *notification.Notification)

	return n
}

func (n *AMQPNotifier) Type() string {
	return "amqp09"
}

func (n *AMQPNotifier) Name() string {
	return n.Config.Name
}

func (n *AMQPNotifier) GetChannel() chan *notification.Notification {
	return n.Channel
}

func (n *AMQPNotifier) Connect() error {
	var err error

	if err = n.wrapper.Dial(n.Address); err != nil {
		return fmt.Errorf("dialing AMQP server: %w", err)
	}

	if err = n.wrapper.Channel(); err != nil {
		return fmt.Errorf("creating channel: %w", err)
	}

	return nil
}

func (n *AMQPNotifier) Close() error {
	if n.wrapper != nil {
		return n.wrapper.CloseConn()
	}

	return fmt.Errorf("can't call Close, Wrapper not set")
}

func (n *AMQPNotifier) Run() {
	for notification := range n.Channel {
		r := n.Deliver(notification)
		if !r.Success {
			n.Logger.Printf("%s: %+v", n.Name(), r)
		}
	}
}

func (n *AMQPNotifier) Notify(payload *notification.Notification) {
	if n.Channel == nil {
		n.Logger.Print("channel is nil")
		return
	}

	if payload == nil {
		n.Logger.Print("payload is nil")
		return
	}

	n.Channel <- payload
}

func (n *AMQPNotifier) Deliver(message *notification.Notification) *notification.Result {
	var (
		ctx, cancel = context.WithTimeout(n.ctx, n.Config.DeliveryTimeout*time.Millisecond)
		err         error
	)

	defer cancel()

	// Serialize the notification data to JSON
	payload, err := n.jsonMarshal(message)
	if err != nil {
		return &notification.Result{Success: false, Error: err}
	}

	err = n.wrapper.PublishWithContext(ctx,
		n.QueueName,                       // exchange
		n.Config.PublishOptions.Key,       // routing key
		n.Config.PublishOptions.Mandatory, // mandatory
		n.Config.PublishOptions.Immediate, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		})
	if err != nil {
		return &notification.Result{Success: false, Error: fmt.Errorf("sending message: %v", err)}
	}

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return &notification.Result{Success: false, Error: fmt.Errorf("message delivery timed out")}
		}
		return &notification.Result{Success: false, Error: ctx.Err()}
	default:
		return &notification.Result{Success: true}
	}
}
