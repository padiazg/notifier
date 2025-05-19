package broker

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/padiazg/notifier/model"
)

// Broker handles the dispatch and tracking of notifications
type Broker struct {
	OnError   func(error)
	notifiers map[string]model.Notifier
}

func New(config *Config) *Broker {
	return (&Broker{}).New(config)
}

func (e *Broker) New(config *Config) *Broker {
	if config == nil {
		config = &Config{}
	}

	if config.OnError != nil {
		e.OnError = config.OnError
	}

	e.notifiers = make(map[string]model.Notifier)

	return e
}

func (e *Broker) NotifierRegister(n model.Notifier) string {
	id := n.Name()
	e.notifiers[id] = n

	return id
}

func (e *Broker) Start() {
	for _, n := range e.notifiers {
		if err := n.Connect(); err != nil {
			e.HandleError(fmt.Errorf("starting notifier %s: %+v", n.Name(), err))
			continue
		}

		go func(n model.Notifier) {
			n.Run()
		}(n)
	}
}

func (e *Broker) Stop() {
	for _, n := range e.notifiers {
		if ch := n.GetChannel(); ch != nil {
			close(ch)
		}
	}
}

func (e *Broker) Dispatch(message *model.Notification) {
	if message == nil {
		return
	}

	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	if len(message.Channels) == 0 {
		e.dispatchAll(message)
	} else {
		e.dispatchChannels(message)
	}
}

func (e *Broker) dispatchAll(message *model.Notification) {
	wg := sync.WaitGroup{}

	for _, n := range e.notifiers {
		fmt.Printf("Broker.dispatchAll %s => (%s) %v\n", n.Name(), message.ID, message.Data)
		wg.Add(1)

		go func(n model.Notifier) {
			defer wg.Done()
			n.Notify(message)
		}(n)
	}

	wg.Wait()
}

func (e *Broker) dispatchChannels(message *model.Notification) {
	wg := sync.WaitGroup{}

	for _, c := range message.Channels {
		n, ok := e.notifiers[c]
		if !ok {
			e.HandleError(fmt.Errorf("%s: channel %s not found", message.ID, c))
			continue
		}

		fmt.Printf("Broker.dispatchChannels %s => (%s) %v\n", n.Name(), message.ID, message.Data)
		wg.Add(1)

		go func(n model.Notifier) {
			defer wg.Done()
			n.Notify(message)
		}(n)
	}
}

func (e *Broker) HandleError(err error) {
	if e.OnError != nil {
		e.OnError(err)
	}
}
