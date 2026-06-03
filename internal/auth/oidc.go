package auth

import (
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

func NewAuthConfig() *AuthConfig {
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcClientId := strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID"))

	cfg := AuthConfig{
		enabled:  true,
		issuer:   oidcIssuer,
		clientId: oidcClientId,
	}

	if len(oidcIssuer) == 0 || len(oidcClientId) == 0 {
		cfg.enabled = false
	}

	if cfg.enabled {
		slog.Info("OIDC authentication enabled with issuer: %s and client ID: %s", cfg.issuer, cfg.clientId)
	}

	return &cfg
}
