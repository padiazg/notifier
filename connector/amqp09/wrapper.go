package amqp09

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqp09WrapperInterface interface {
	Dial(url string) error
	CloseConn() error
	Channel() error
	CloseChannel() error
	PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
}

type amqp09Wrapper struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (w *amqp09Wrapper) Dial(url string) (err error) {
	w.conn, err = amqp.Dial(url)
	return err
}

func (w *amqp09Wrapper) CloseConn() error {
	return w.conn.Close()
}

func (w *amqp09Wrapper) Channel() (err error) {
	w.channel, err = w.conn.Channel()
	return err
}

func (w *amqp09Wrapper) CloseChannel() error {
	return w.channel.Close()
}

func (w *amqp09Wrapper) PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	return w.channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}
