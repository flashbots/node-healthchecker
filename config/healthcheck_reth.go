package config

import (
	"fmt"
	"net/url"
	"time"
)

type HealthcheckReth struct {
	BaseURL           string        `yaml:"base_url"`
	BlockAgeThreshold time.Duration `yaml:"-"`
}

func (c *HealthcheckReth) Preprocess() error {
	if c.BaseURL != "" {
		if _, err := url.Parse(c.BaseURL); err != nil {
			return fmt.Errorf("invalid reth base url: %w",
				err,
			)
		}
	}
	return nil
}
