package notification

// EventType represents the possible event types
type EventType string

// Notification represents a notification with its details
type Notification struct {
	ID    string
	Event EventType
	Data  interface{}
}

// Result represents the result of sending a notification
type Result struct {
	Success bool
	Error   error
}
