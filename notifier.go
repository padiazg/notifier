package notifier

// EventType represents the possible event types
type EventType string

// Notification represents a notification with its details
type Notification struct {
	ID    string
	Event EventType
	Data  interface{}
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	Success bool
	Error   error
}

// Notifier is the interface for sending notifications
type Notifier interface {
	SendNotification(notification *Notification) NotificationResult
}
