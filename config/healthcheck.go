package config

import "time"

type Healthcheck struct {
	CacheTimeout time.Duration `yaml:"cache_timeout"`
	Timeout      time.Duration `yaml:"timeout"`
}
