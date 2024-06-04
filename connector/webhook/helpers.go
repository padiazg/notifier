package webhook

func NewWebhookNotifier(config *Config) *WebhookNotifier {
	return (&WebhookNotifier{}).New(config)
}
