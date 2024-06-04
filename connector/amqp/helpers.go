package amqp

import (
	n "github.com/padiazg/notifier/notification"
)

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
