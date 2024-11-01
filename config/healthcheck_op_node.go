package config

import (
	"fmt"
	"net/url"
	"time"
)

type HealthcheckOpNode struct {
	BaseURL              string        `yaml:"base_url"`
	BlockAgeThreshold    time.Duration `yaml:"-"`
	ConfirmationDistance uint64        `yaml:"confirmation_distance"`
}

func (c *HealthcheckOpNode) Preprocess() error {
	if c.BaseURL != "" {
		if _, err := url.Parse(c.BaseURL); err != nil {
			return fmt.Errorf("invalid op-node base url: %w",
				err,
			)
		}
	}
	return nil
}
