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
	jwksURL  string
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
func (cfg *AuthConfig) JWKSURL() string {
	return cfg.jwksURL
}

func NewAuthConfig() (*AuthConfig, error) {
	cfg := AuthConfig{
		enabled:  false,
		jwksURL:  "",
		issuer:   "",
		audience: "",
	}
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcAudience := strings.TrimSpace(os.Getenv("OIDC_AUDIENCE"))
	oidcJWKSURL := strings.TrimSpace(os.Getenv("OIDC_JWKS_URL"))

	if oidcIssuer == "" && oidcAudience == "" && oidcJWKSURL == "" {
		cfg.enabled = false
		return &cfg, nil
	}

	lengths := []int{len(oidcIssuer), len(oidcAudience), len(oidcJWKSURL)}

	partialFound := 0
	for _, l := range lengths {
		if l > 0 {
			partialFound++
		}
	}

	if partialFound == 3 {
		cfg.enabled = true
		cfg.issuer = oidcIssuer
		cfg.audience = oidcAudience
		cfg.jwksURL = oidcJWKSURL
	} else if partialFound < 3 {
		errMsg := "OIDC_ISSUER, OIDC_AUDIENCE, and OIDC_JWKS_URL must be set"
		slog.Error(errMsg, slog.String("OIDC_ISSUER", oidcIssuer), slog.String("OIDC_AUDIENCE", oidcAudience), slog.String("OIDC_JWKS_URL", oidcJWKSURL))
		return nil, errors.New(errMsg)
	}

	return &cfg, nil
}
