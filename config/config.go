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
