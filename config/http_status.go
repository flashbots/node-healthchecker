package config

type HttpStatus struct {
	Ok      int `yaml:"ok"`
	Warning int `yaml:"warning"`
	Error   int `yaml:"error"`
}
