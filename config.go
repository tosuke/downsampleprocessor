package downsampleprocessor

import (
	"time"

	"go.opentelemetry.io/collector/component"
)

type Config struct {
	Duration       time.Duration `mapstructure:"duration"`
	MaxCardinality int           `mapstructure:"max_cardinality"`
}

var _ component.ConfigValidator = (*Config)(nil)

func (c *Config) Validate() error {
	return nil
}

func createDefaultConfig() component.Config {
	return &Config{}
}
