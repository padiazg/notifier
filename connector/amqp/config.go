package amqp

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
