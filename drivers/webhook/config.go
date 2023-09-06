package webhook

type Config struct {
	Endpoint string
	Insecure bool
	Headers  map[string]string
}
