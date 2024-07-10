package model

// Notifier is the interface for sending notifications
type Notifier interface {
	Type() string
	Name() string
	Connect() error
	Close() error
	Run()
	GetChannel() chan *Notification
	Notify(notification *Notification)
	Deliver(notification *Notification) *Result
}
