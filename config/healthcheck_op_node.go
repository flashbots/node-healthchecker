package config

type HealthcheckOpNode struct {
	BaseURL              string `yaml:"base_url"`
	ConfirmationDistance uint64 `yaml:"confirmation_distance"`
}
