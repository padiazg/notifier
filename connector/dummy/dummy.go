package dummy

import (
	"fmt"
	"sync"

	"github.com/padiazg/notifier/notification"
)

type Config struct {
	ConnectError error
	Name         string
}

type DummyNotifier struct {
	*Config
	Channel chan *notification.Notification
	in      []*notification.Notification
	lock    *sync.RWMutex
}

func New(config *Config) *DummyNotifier {
	dn := &DummyNotifier{
		in:   make([]*notification.Notification, 0),
		lock: &sync.RWMutex{},
	}

	if config != nil {
		dn.Config = config
	}

	return dn
}

func (d *DummyNotifier) Connect() error {
	if d.ConnectError != nil {
		return d.ConnectError
	}

	return nil
}

func (d *DummyNotifier) Close() error {
	return nil
}

func (d *DummyNotifier) StartWorker() {
	d.Channel = make(chan *notification.Notification)
	for notification := range d.Channel {
		d.SendNotification(notification)
	}
}

func (d *DummyNotifier) GetChannel() chan *notification.Notification {
	return d.Channel
}

func (d *DummyNotifier) Notify(payload *notification.Notification) {
	if d.Channel == nil || payload == nil {
		return
	}

	d.Channel <- payload
}

func (d *DummyNotifier) SendNotification(message *notification.Notification) *notification.Result {
	d.lock.Lock()
	d.in = append(d.in, message)
	defer d.lock.Unlock()

	res, ok := message.Data.(*notification.Result)
	if !ok {
		res = &notification.Result{Error: fmt.Errorf("unexpected type: %T", message.Data)}
	}
	return res
}

func (d *DummyNotifier) Name() string                     { return d.Config.Name }
func (d *DummyNotifier) Type() string                     { return "dummy" }
func (d *DummyNotifier) In() []*notification.Notification { return d.in }

func (d *DummyNotifier) Exists(n *notification.Notification) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, data := range d.in {
		if data == n {
			return true
		}
	}
	return false
}

func (d *DummyNotifier) First() *notification.Notification {
	d.lock.Lock()
	defer d.lock.Unlock()
	for _, data := range d.in {
		return data
	}
	return nil
}
