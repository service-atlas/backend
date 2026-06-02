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
	clientId string
}

func (cfg *AuthConfig) Enabled() bool {
	return cfg.enabled
}

func (cfg *AuthConfig) Issuer() string {
	return cfg.issuer
}
func (cfg *AuthConfig) ClientId() string {
	return cfg.clientId
}

func NewAuthConfig() (*AuthConfig, error) {
	cfg := AuthConfig{
		enabled:  false,
		issuer:   "",
		clientId: "",
	}
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcClientId := strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID"))

	if oidcIssuer == "" && oidcClientId == "" {
		cfg.enabled = false
		return &cfg, nil
	}

	if len(oidcIssuer) > 0 && len(oidcClientId) > 0 {
		cfg.enabled = true
		cfg.issuer = oidcIssuer
		cfg.clientId = oidcClientId
	} else {
		errMsg := "OIDC_ISSUER, OIDC_AUDIENCE, and OIDC_CLIENT_ID must be set"
		slog.Error(errMsg, slog.String("OIDC_ISSUER", oidcIssuer), slog.String("OIDC_AUDIENCE", oidcClientId))
		return nil, errors.New(errMsg)
	}

	return &cfg, nil
}
