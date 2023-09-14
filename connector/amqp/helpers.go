package amqp

import (
	n "github.com/padiazg/notifier/notification"
)

func NewAMQPNotifier(config *Config) n.Notifier {
	var notifier n.Notifier

	switch config.Protocol {
	// case ProtocolAMQP09:
	// 	notifier = NewAMQP09Notifier(config)
	case ProtocolAMQP10:
		notifier = NewAMQP10Notifier(config)
	}

	return notifier
}

func NewAMQP10Notifier(config *Config) *AMQP10Notifier {
	return &AMQP10Notifier{
		Config: config,
	}
}

func NewAMQP09Notifier(config *Config) *AMQP09Notifier {
	return &AMQP09Notifier{
		Config: config,
	}
}
