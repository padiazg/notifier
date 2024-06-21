package dummy

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/padiazg/notifier/notification"
)

type Config struct {
	Logger       *log.Logger
	ConnectError error
	Name         string
}

type DummyNotifier struct {
	lock *sync.RWMutex
	*Config
	Channel chan *notification.Notification
	in      []*notification.Notification
}

func New(config *Config) *DummyNotifier {
	n := &DummyNotifier{
		in:   make([]*notification.Notification, 0),
		lock: &sync.RWMutex{},
	}

	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = "dummya0b1c3d4"
	}

	n.Config = config

	n.Channel = make(chan *notification.Notification)

	if config.Logger == nil {
		config.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return n
}

func (n *DummyNotifier) Connect() error {
	if n.ConnectError != nil {
		return n.ConnectError
	}

	return nil
}

func (n *DummyNotifier) Close() error {
	return nil
}

func (n *DummyNotifier) StartWorker() {
	n.Channel = make(chan *notification.Notification)
	for notification := range n.Channel {
		n.SendNotification(notification)
	}
}

func (n *DummyNotifier) GetChannel() chan *notification.Notification {
	return n.Channel
}

func (n *DummyNotifier) Notify(payload *notification.Notification) {
	if n.Channel == nil {
		n.Logger.Print("channel is nil")
		return
	}

	if payload == nil {
		n.Logger.Print("payload is nil")
		return
	}

	n.Channel <- payload
}

func (n *DummyNotifier) SendNotification(message *notification.Notification) *notification.Result {
	n.lock.Lock()
	n.in = append(n.in, message)
	defer n.lock.Unlock()

	res, ok := message.Data.(*notification.Result)
	if !ok {
		res = &notification.Result{Error: fmt.Errorf("unexpected type: %T", message.Data)}
	}

	return res
}

func (n *DummyNotifier) Name() string                     { return n.Config.Name }
func (n *DummyNotifier) Type() string                     { return "dummy" }
func (n *DummyNotifier) In() []*notification.Notification { return n.in }

func (n *DummyNotifier) Exists(item *notification.Notification) bool {
	n.lock.Lock()

	defer n.lock.Unlock()

	for _, data := range n.in {
		if data == item {
			return true
		}
	}

	return false
}

func (n *DummyNotifier) First() *notification.Notification {
	n.lock.Lock()
	defer n.lock.Unlock()

	for _, data := range n.in {
		return data
	}

	return nil
}
