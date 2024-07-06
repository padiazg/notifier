package amqp09

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type internalWrapperInterface interface {
	Dial(url string) error
	CloseConn() error
	Channel() error
	CloseChannel() error
	PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
}

type internalWrapper struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (w *internalWrapper) Dial(url string) error {
	var err error
	w.conn, err = amqp.Dial(url)
	return err
}

func (w *internalWrapper) CloseConn() error {
	return w.conn.Close()
}

func (w *internalWrapper) Channel() error {
	var err error
	w.channel, err = w.conn.Channel()
	return err
}

func (w *internalWrapper) CloseChannel() error {
	return w.channel.Close()
}

func (w *internalWrapper) PublishWithContext(ctx context.Context, exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	return w.channel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}
