package config

import (
	"fmt"
	"net/url"
	"time"
)

type HealthcheckLighthouse struct {
	BaseURL           string        `yaml:"base_url"`
	BlockAgeThreshold time.Duration `yaml:"-"`
}

func (c *HealthcheckLighthouse) Preprocess() error {
	if c.BaseURL != "" {
		if _, err := url.Parse(c.BaseURL); err != nil {
			return fmt.Errorf("invalid lighthouse base url: %w",
				err,
			)
		}
	}
	return nil
}
