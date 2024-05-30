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
	notifiers []notification.Notifier
}

func (e *Engine) AddNotifier(n notification.Notifier) {
	e.notifiers = append(e.notifiers, n)
}

func (e *Engine) Start() error {
	fmt.Println("Engine.Start starting workers...")

	for _, n := range e.notifiers {
		fmt.Printf("Starting worker %s\n", n.Name())
		if err := n.Connect(); err != nil {
			fmt.Printf("Error starting worker %T: %v\n", n.Name(), err)
			continue
		}

		go func(n notification.Notifier) {
			n.StartWorker()
		}(n)
	}

	return nil
}

func (e *Engine) Stop() error {
	fmt.Println("Engine.Stop")

	for _, n := range e.notifiers {
		fmt.Printf("Stopping worker %T\n", n.Name())
		if ch := n.GetChannel(); ch != nil {
			close(ch)
		}
	}

	return nil
}

func (e *Engine) Dispatch(message *notification.Notification) {
	fmt.Printf("Engine.Dispatch\n")
	if message == nil {
		return
	}

	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	wg := sync.WaitGroup{}

	for _, n := range e.notifiers {
		wg.Add(1)
		go func(n notification.Notifier) {
			defer wg.Done()
			n.Notify(message)
		}(n)
	}

	wg.Wait()
}

func (e *Engine) HandleError(err error) {
	if e.OnError != nil {
		e.OnError(err)
	}
}
