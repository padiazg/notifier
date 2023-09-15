package webhook

type Config struct {
	Name     string
	Endpoint string
	Insecure bool
	Headers  map[string]string
}
