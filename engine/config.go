package engine

import (
	ad "github.com/padiazg/notifier/drivers/amqp"
	wd "github.com/padiazg/notifier/drivers/webhook"
)

type Config struct {
	WebHook *wd.Config
	MQ      *ad.Config
	OnError func(error)
}
