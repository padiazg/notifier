package notifier

// MQNotifier implements the Notifier interface for message queues
type MQNotifier struct {
	// You can add fields specific to the MQNotifier
}

func (mn *MQNotifier) SendNotification(notification Notification) NotificationResult {
	// Implement MQ notification logic here
	return NotificationResult{Success: true}
}
