package downsampleprocessor

import (
	"time"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	Period         time.Duration `mapstructure:"period"`
	MaxCardinality int           `mapstructure:"max_cardinality"`
}

var _ component.ConfigValidator = (*Config)(nil)

func (c *Config) Validate() error {
	return nil
}

func createDefaultConfig() component.Config {
	return &Config{}
}
