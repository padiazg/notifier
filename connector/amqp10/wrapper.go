package amqp10

import (
	"context"

	amqp "github.com/Azure/go-amqp"
)

type amqp10WrapperInterface interface {
	Dial(ctx context.Context, addr string, opts *amqp.ConnOptions) error
	NewSession(ctx context.Context, opts *amqp.SessionOptions) error
	NewSender(ctx context.Context, target string, opts *amqp.SenderOptions) error
	Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error
	CloseConn() error
	CloseSession(ctx context.Context) error
}

type amqp10Wrapper struct {
	conn    *amqp.Conn
	session *amqp.Session
	sender  *amqp.Sender
}

func (w *amqp10Wrapper) Dial(ctx context.Context, addr string, opts *amqp.ConnOptions) error {
	var err error
	w.conn, err = amqp.Dial(ctx, addr, opts)
	return err
}

func (w *amqp10Wrapper) NewSession(ctx context.Context, opts *amqp.SessionOptions) error {
	var err error
	w.session, err = w.conn.NewSession(ctx, opts)
	return err
}

func (w *amqp10Wrapper) NewSender(ctx context.Context, target string, opts *amqp.SenderOptions) error {
	var err error
	w.sender, err = w.session.NewSender(ctx, target, opts)
	return err
}

func (w *amqp10Wrapper) CloseConn() error {
	return w.conn.Close()
}

func (w *amqp10Wrapper) CloseSession(ctx context.Context) error {
	return w.session.Close(ctx)
}

func (w *amqp10Wrapper) Send(ctx context.Context, msg *amqp.Message, opts *amqp.SendOptions) error {
	return w.sender.Send(ctx, msg, opts)
}
