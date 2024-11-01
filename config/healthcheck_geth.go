package config

import (
	"fmt"
	"net/url"
	"time"
)

type HealthcheckGeth struct {
	BaseURL           string        `yaml:"base_url"`
	BlockAgeThreshold time.Duration `yaml:"-"`
}

func (c *HealthcheckGeth) Preprocess() error {
	if c.BaseURL != "" {
		if _, err := url.Parse(c.BaseURL); err != nil {
			return fmt.Errorf("invalid geth base url: %w",
				err,
			)
		}
	}
	return nil
}
