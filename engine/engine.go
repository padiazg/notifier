package engine

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/padiazg/notifier/notification"
)

// Engine handles the dispatch and tracking of notifications
type Engine struct {
	OnError   func(error)
	notifiers map[string]notification.Notifier
}

func NewEngine(config *Config) *Engine {
	return (&Engine{}).New(config)
}

func (e *Engine) New(config *Config) *Engine {
	if config == nil {
		config = &Config{}
	}

	if config.OnError != nil {
		e.OnError = config.OnError
	}

	e.notifiers = make(map[string]notification.Notifier)

	return e
}

func (e *Engine) RegisterNotifier(n notification.Notifier) string {
	id := n.Name()
	e.notifiers[id] = n

	return id
}

func (e *Engine) Start() {
	for _, n := range e.notifiers {
		if err := n.Connect(); err != nil {
			e.HandleError(fmt.Errorf("starting notifier %s: %+v", n.Name(), err))
			continue
		}

		go func(n notification.Notifier) {
			n.Run()
		}(n)
	}
}

func (e *Engine) Stop() {
	for _, n := range e.notifiers {
		if ch := n.GetChannel(); ch != nil {
			close(ch)
		}
	}
}

func (e *Engine) Dispatch(message *notification.Notification) {
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

func (e *Engine) dispatchAll(message *notification.Notification) {
	wg := sync.WaitGroup{}

	for _, n := range e.notifiers {
		fmt.Printf("Engine.dispatchAll %s => (%s) %v\n", n.Name(), message.ID, message.Data)
		wg.Add(1)

		go func(n notification.Notifier) {
			defer wg.Done()
			n.Notify(message)
		}(n)
	}

	wg.Wait()
}

func (e *Engine) dispatchChannels(message *notification.Notification) {
	wg := sync.WaitGroup{}

	for _, c := range message.Channels {
		n, ok := e.notifiers[c]
		if !ok {
			e.HandleError(fmt.Errorf("%s: channel %s not found", message.ID, c))
			continue
		}

		fmt.Printf("Engine.dispatchChannels %s => (%s) %v\n", n.Name(), message.ID, message.Data)
		wg.Add(1)

		go func(n notification.Notifier) {
			defer wg.Done()
			n.Notify(message)
		}(n)
	}
}

func (e *Engine) HandleError(err error) {
	if e.OnError != nil {
		e.OnError(err)
	}
}
