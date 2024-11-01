package config

type Server struct {
	ListenAddress string `yaml:"listen_address"`
}

func (c *Server) Preprocess() error {
	return nil
}
