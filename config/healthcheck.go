package config

import "time"

type Healthcheck struct {
	Timeout time.Duration `yaml:"timeout"`
}
