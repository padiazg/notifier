package amqp10

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/Azure/go-amqp"
	"github.com/padiazg/notifier/model"
	"github.com/padiazg/notifier/utils"
)

type Config struct {
	Logger          *log.Logger
	Name            string
	QueueName       string
	Address         string
	DeliveryTimeout time.Duration
	SessionOptions  *amqp.SessionOptions
	ConnOptions     *amqp.ConnOptions
	SenderOptions   *amqp.SenderOptions
	SendOptions     *amqp.SendOptions
	wrapper         internalWrapperInterface
	ctx             context.Context
}

// AMQPNotifier implements the Notifier interface for message queues
type AMQPNotifier struct {
	*Config
	Channel     chan *model.Notification
	jsonMarshal func(v any) ([]byte, error)
}

var _ model.Notifier = (*AMQPNotifier)(nil)

func New(config *Config) *AMQPNotifier {
	return (&AMQPNotifier{}).New(config)
}

func (n *AMQPNotifier) New(config *Config) *AMQPNotifier {
	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = n.Type() + utils.RandomId(utils.ID8)
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
	n.Channel = make(chan *model.Notification)

	return n
}

func (n *AMQPNotifier) Type() string {
	return "amqp10"
}

func (n *AMQPNotifier) Name() string {
	return n.Config.Name
}

func (n *AMQPNotifier) GetChannel() chan *model.Notification {
	return n.Channel
}

func (n *AMQPNotifier) Connect() error {
	var err error

	// create a connection
	if err = n.wrapper.Dial(n.ctx, n.Address, n.ConnOptions); err != nil {
		return fmt.Errorf("dialing AMQP server: %w", err)
	}

	// create a session
	if err = n.wrapper.NewSession(n.ctx, n.SessionOptions); err != nil {
		return fmt.Errorf("creating AMQP session: %w", err)
	}

	// create a sender
	if err = n.wrapper.NewSender(n.ctx, n.QueueName, n.SenderOptions); err != nil {
		return fmt.Errorf("creating sender link: %w", err)
	}

	return nil
}

func (n *AMQPNotifier) Close() error {
	if n.wrapper != nil {
		return n.wrapper.CloseConn()
	}

	return fmt.Errorf("can't call Close, Wrapper not set")
}

func (n *AMQPNotifier) Notify(payload *model.Notification) {
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

func (n *AMQPNotifier) Run() {
	for notification := range n.Channel {
		r := n.Deliver(notification)
		if !r.Success {
			n.Logger.Printf("%s: %+v\n", n.Name(), r)
		}
	}
}

func (n *AMQPNotifier) Deliver(message *model.Notification) *model.Result {
	var (
		ctx, cancel = context.WithTimeout(n.ctx, n.Config.DeliveryTimeout*time.Millisecond)
		err         error
	)

	defer cancel()

	// Serialize the notification data to JSON
	payload, err := n.jsonMarshal(message)
	if err != nil {
		return &model.Result{Success: false, Error: err}
	}

	// send message
	err = n.wrapper.Send(ctx, amqp.NewMessage(payload), n.SendOptions)
	if err != nil {
		return &model.Result{Success: false, Error: fmt.Errorf("sending message: %v", err)}
	}

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return &model.Result{Success: false, Error: fmt.Errorf("message delivery timed out")}
		}
		return &model.Result{Success: false, Error: ctx.Err()}
	default:
		return &model.Result{Success: true}
	}
}
