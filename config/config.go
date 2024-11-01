package config

type Config struct {
	Log    Log    `yaml:"log"`
	Server Server `yaml:"server"`

	HttpStatus HttpStatus `yaml:"http_status"`

	Healthcheck Healthcheck `yaml:"healthcheck"`

	HealthcheckGeth       HealthcheckGeth       `yaml:"healthcheck_geth"`
	HealthcheckLighthouse HealthcheckLighthouse `yaml:"healthcheck_lighthouse"`
	HealthcheckOpNode     HealthcheckOpNode     `yaml:"healthcheck_op_node"`
	HealthcheckReth       HealthcheckReth       `yaml:"healthcheck_reth"`
}

func (c *Config) Preprocess() error {
	errs := make([]error, 0)

	if c.Healthcheck.BlockAgeThreshold != 0 {
		c.HealthcheckGeth.BlockAgeThreshold = c.Healthcheck.BlockAgeThreshold
		c.HealthcheckLighthouse.BlockAgeThreshold = c.Healthcheck.BlockAgeThreshold
		c.HealthcheckOpNode.BlockAgeThreshold = c.Healthcheck.BlockAgeThreshold
		c.HealthcheckReth.BlockAgeThreshold = c.Healthcheck.BlockAgeThreshold
	}

	errs = append(errs, c.Log.Preprocess())
	errs = append(errs, c.Server.Preprocess())
	errs = append(errs, c.HttpStatus.Preprocess())
	errs = append(errs, c.Healthcheck.Preprocess())
	errs = append(errs, c.HealthcheckGeth.Preprocess())
	errs = append(errs, c.HealthcheckLighthouse.Preprocess())
	errs = append(errs, c.HealthcheckOpNode.Preprocess())
	errs = append(errs, c.HealthcheckReth.Preprocess())

	return flatten(errs)
}
