package kafka

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/padiazg/notifier/model"
	"github.com/padiazg/notifier/utils"
)

type Config struct {
	Logger       *log.Logger
	ConnectError error
	Name         string
	ctx          context.Context
}

type KafkaNotifier struct {
	*Config
	channel chan *model.Notification
}

var _ model.Notifier = (*KafkaNotifier)(nil)

func (n *KafkaNotifier) New(config *Config) *KafkaNotifier {
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

	return n
}

func (n *KafkaNotifier) Type() string {
	return "kafka"
}

func (n *KafkaNotifier) Name() string {
	return n.Config.Name
}

func (n *KafkaNotifier) GetChannel() chan *model.Notification {
	return n.channel
}

func (n *KafkaNotifier) Connect() error {
	// var err error

	// create a connection

	// return nil
	return fmt.Errorf("%s notifier not implemented", n.Type())
}

func (n *KafkaNotifier) Close() error {
	// if n.wrapper != nil {
	// 	return n.wrapper.CloseConn()
	// }

	return fmt.Errorf("can't call Close, Wrapper not set")
}

func (n *KafkaNotifier) Notify(payload *model.Notification) {
	if n.channel == nil {
		n.Logger.Print("channel is nil")
		return
	}

	if payload == nil {
		n.Logger.Print("payload is nil")
		return
	}

	n.channel <- payload
}

func (n *KafkaNotifier) Run() {
	for notification := range n.channel {
		r := n.Deliver(notification)
		if !r.Success {
			n.Logger.Printf("%s: %+v\n", n.Name(), r)
		}
	}
}

func (n *KafkaNotifier) Deliver(message *model.Notification) *model.Result {
	// var (
	// 	ctx, cancel = context.WithTimeout(n.ctx, n.Config.DeliveryTimeout*time.Millisecond)
	// 	err         error
	// )

	// defer cancel()

	// // Serialize the notification data to JSON
	// payload, err := n.jsonMarshal(message)
	// if err != nil {
	// 	return &model.Result{Success: false, Error: err}
	// }

	// // send message
	// err = n.wrapper.Send(ctx, amqp.NewMessage(payload), n.SendOptions)
	// if err != nil {
	// 	return &model.Result{Success: false, Error: fmt.Errorf("sending message: %v", err)}
	// }

	// select {
	// case <-ctx.Done():
	// 	if ctx.Err() == context.DeadlineExceeded {
	// 		return &model.Result{Success: false, Error: fmt.Errorf("message delivery timed out")}
	// 	}
	// 	return &model.Result{Success: false, Error: ctx.Err()}
	// default:
	return &model.Result{Success: true}
	// }
}
