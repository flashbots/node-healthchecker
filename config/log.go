package config

type Log struct {
	Level string `yaml:"level"`
	Mode  string `yaml:"mode"`
}

func (c *Log) Preprocess() error {
	return nil
}
