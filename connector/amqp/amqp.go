package amqp

import (
	n "github.com/padiazg/notifier/notification"
)

type Protocol uint

const (
	ProtocolAMQP09 Protocol = iota
	ProtocolAMQP10
)

type Config struct {
	Name      string
	QueueName string
	Address   string
	Protocol  Protocol
}

func NewAMQPNotifier(config *Config) n.Notifier {
	var notifier n.Notifier

	switch config.Protocol {
	case ProtocolAMQP09:
		notifier = (&AMQP10Notifier{}).New(config)
	case ProtocolAMQP10:
		notifier = (&AMQP09Notifier{}).New(config)
	}

	return notifier
}
