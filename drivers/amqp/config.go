package amqp

import n "github.com/padiazg/notifier/notification"

type Protocol uint

const (
	ProtocolAMQP09 Protocol = iota
	ProtocolAMQP10
)

type Config struct {
	QueueName string
	Address   string
	Protocol  Protocol
}

func NewAMQPNotifier(config *Config) n.Notifier {
	var notifier n.Notifier

	switch config.Protocol {
	case ProtocolAMQP09:
		notifier = NewAMQP09Notifier(config)
	case ProtocolAMQP10:
		notifier = NewAMQP10Notifier(config)
	}

	return notifier
}
