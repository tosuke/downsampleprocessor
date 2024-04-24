package downsampleprocessor

import "go.opentelemetry.io/collector/component"

type Config struct{}

var _ component.ConfigValidator = (*Config)(nil)

func (c *Config) Validate() error {
	return nil
}

func createDefaultConfig() component.Config {
	return &Config{}
}
