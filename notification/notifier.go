package notification

// Notifier is the interface for sending notifications
type Notifier interface {
	Connect() error
	Close() error
	StartWorker()
	GetChannel() chan *Notification
	Notify(notification *Notification)
	SendNotification(notification *Notification) Result
	Name() string
}
