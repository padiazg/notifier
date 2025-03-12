package webhook

func NewWebhookNotifier(config *Config) *WebhookNotifier {
	return &WebhookNotifier{
		Config: config,
	}
}
