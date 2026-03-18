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
	ApplicationPath string `env:"APP_PATH" envDefault:"./printyl"`
	Port            uint   `env:"PORT" envDefault:"8080"`
}
