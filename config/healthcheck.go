package config

import "time"

type Healthcheck struct {
	BlockAgeThreshold         time.Duration `yaml:"block_age_threshold"`
	CacheCoolOff              time.Duration `yaml:"cache_cool_off"`
	Timeout                   time.Duration `yaml:"timeout"`
	UnconditionalFailDuration time.Duration `yaml:"unconditional_fail_duration"`
}

func (c *Healthcheck) Preprocess() error {
	return nil
}
