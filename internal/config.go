package internal

import "github.com/caarlos0/env"

var Cfg *Config

func LoadConfig() error {
	Cfg = &Config{}
	if err := env.Parse(Cfg); err != nil {
		return err
	}

	return nil
}

type Config struct {
	DocumentsPath string `env:"DOCUMENTS_PATH" envDefault:"./printyl/documents"`
	Port          uint   `env:"PORT" envDefault:"8080"`
}
