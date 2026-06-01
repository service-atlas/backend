package auth

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

type AuthConfig struct {
	enabled  bool
	issuer   string
	audience string
}

func (cfg *AuthConfig) Enabled() bool {
	return cfg.enabled
}

func (cfg *AuthConfig) Issuer() string {
	return cfg.issuer
}
func (cfg *AuthConfig) Audience() string {
	return cfg.audience
}

func NewAuthConfig() (*AuthConfig, error) {
	cfg := AuthConfig{
		enabled:  false,
		issuer:   "",
		audience: "",
	}
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcAudience := strings.TrimSpace(os.Getenv("OIDC_AUDIENCE"))

	if oidcIssuer == "" && oidcAudience == "" {
		cfg.enabled = false
		return &cfg, nil
	}

	if len(oidcIssuer) > 0 && len(oidcAudience) > 0 {
		cfg.enabled = true
		cfg.issuer = oidcIssuer
		cfg.audience = oidcAudience
	} else {
		errMsg := "OIDC_ISSUER, OIDC_AUDIENCE, and OIDC_JWKS_URL must be set"
		slog.Error(errMsg, slog.String("OIDC_ISSUER", oidcIssuer), slog.String("OIDC_AUDIENCE", oidcAudience))
		return nil, errors.New(errMsg)
	}

	return &cfg, nil
}
