package dummy

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/padiazg/notifier/model"
)

type Config struct {
	Logger       *log.Logger
	ConnectError error
	Name         string
}

type DummyNotifier struct {
	lock *sync.RWMutex
	*Config
	Channel chan *model.Notification
	in      []*model.Notification
}

var _ model.Notifier = (*DummyNotifier)(nil)

func New(config *Config) *DummyNotifier {
	n := &DummyNotifier{
		in:   make([]*model.Notification, 0),
		lock: &sync.RWMutex{},
	}

	if config == nil {
		config = &Config{}
	}

	if config.Name == "" {
		config.Name = "dummya0b1c3d4"
	}

	n.Config = config

	n.Channel = make(chan *model.Notification)

	if config.Logger == nil {
		config.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return n
}

func (n *DummyNotifier) Type() string { return "dummy" }
func (n *DummyNotifier) Name() string { return n.Config.Name }

func (n *DummyNotifier) Connect() error {
	if n.ConnectError != nil {
		return n.ConnectError
	}

	return nil
}

func (n *DummyNotifier) Close() error {
	return nil
}

func (n *DummyNotifier) Run() {
	n.Channel = make(chan *model.Notification)
	for notification := range n.Channel {
		r := n.Deliver(notification)
		if !r.Success {
			n.Logger.Printf("%s: %+v", n.Name(), r)
		}
	}
}

func (n *DummyNotifier) GetChannel() chan *model.Notification {
	return n.Channel
}

func (n *DummyNotifier) Notify(payload *model.Notification) {
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

func (n *DummyNotifier) Deliver(message *model.Notification) *model.Result {
	n.lock.Lock()
	n.in = append(n.in, message)
	defer n.lock.Unlock()

	// to trigger a deliver error assign message.Data a value different than model.Result
	res, ok := message.Data.(*model.Result)
	if !ok {
		res = &model.Result{Error: fmt.Errorf("unexpected type: %T", message.Data)}
	}

	return res
}

func (n *DummyNotifier) In() []*model.Notification { return n.in }

func (n *DummyNotifier) Exists(item *model.Notification) bool {
	n.lock.Lock()

	defer n.lock.Unlock()

	for _, data := range n.in {
		if data == item {
			return true
		}
	}

	return false
}

func (n *DummyNotifier) First() *model.Notification {
	n.lock.Lock()
	defer n.lock.Unlock()

	for _, data := range n.in {
		return data
	}

	return nil
}
