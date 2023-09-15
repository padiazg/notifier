package engine

func NewEngine(config *Config) *Engine {
	engine := &Engine{}

	if config == nil {
		return engine
	}

	if config.OnError != nil {
		engine.OnError = config.OnError
	}

	return engine
}
