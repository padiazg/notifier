module github.com/padiazg/notifier/example

go 1.20

require github.com/padiazg/notifier v0.0.0-00010101000000-000000000000

require (
	github.com/Azure/go-amqp v1.0.1 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/rabbitmq/amqp091-go v1.8.1 // indirect
)

replace github.com/padiazg/notifier => ../
