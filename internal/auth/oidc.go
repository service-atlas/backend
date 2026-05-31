package auth

import (
	"errors"
	"log/slog"
	"os"
	"strings"
)

type AuthConfig struct {
	Enabled  bool
	Issuer   string
	Audience string
	JWKSURL  string
}

func NewAuthConfig() (*AuthConfig, error) {
	cfg := AuthConfig{
		Enabled:  false,
		JWKSURL:  "",
		Issuer:   "",
		Audience: "",
	}
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcAudience := strings.TrimSpace(os.Getenv("OIDC_AUDIENCE"))
	oidcJWKSURL := strings.TrimSpace(os.Getenv("OIDC_JWKS_URL"))

	if oidcIssuer == "" && oidcAudience == "" && oidcJWKSURL == "" {
		cfg.Enabled = false
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
		cfg.Enabled = true
		cfg.Issuer = oidcIssuer
		cfg.Audience = oidcAudience
		cfg.JWKSURL = oidcJWKSURL
	} else if partialFound < 3 {
		errMsg := "OIDC_ISSUER, OIDC_AUDIENCE, and OIDC_JWKS_URL must be set"
		slog.Error(errMsg, slog.String("OIDC_ISSUER", oidcIssuer), slog.String("OIDC_AUDIENCE", oidcAudience), slog.String("OIDC_JWKS_URL", oidcJWKSURL))
		return nil, errors.New(errMsg)
	}

	return &cfg, nil
}
