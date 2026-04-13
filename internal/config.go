package internal

import (
	"fmt"

	"github.com/caarlos0/env"
)

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
	LatexImage      string `env:"LATEX_IMAGE" envDefault:"texlive/texlive:latest"`

	OIDCIssuerURL      string   `env:"OIDC_ISSUER_URL"`
	OIDCClientID       string   `env:"OIDC_CLIENT_ID"`
	AuthEnabled        bool     `env:"AUTH_ENABLED" envDefault:"false"`
	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:","`
}

func (cfg *Config) OICDEnabled() (bool, error) {
	if !cfg.AuthEnabled {
		return false, nil
	}

	if len(cfg.CORSAllowedOrigins) == 0 {
		return false, fmt.Errorf("CORS_ALLOWED_ORIGINS is empty")
	}

	if cfg.OIDCIssuerURL == "" {
		return false, fmt.Errorf("OIDC_ISSUER_URL is empty")
	}

	if cfg.OIDCClientID == "" {
		return false, fmt.Errorf("OIDC_CLIENT_ID is empty")
	}

	return true, nil
}
