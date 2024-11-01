package config

type HttpStatus struct {
	Ok      int `yaml:"ok"`
	Warning int `yaml:"warning"`
	Error   int `yaml:"error"`
}

func (c *HttpStatus) Preprocess() error {
	return nil
}
