package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

type AuthConfig struct {
	enabled  bool
	Verifier TokenVerifier
}

func (cfg *AuthConfig) Enabled() bool {
	return cfg.enabled
}

func NewAuthConfig() (*AuthConfig, error) {
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcClientId := strings.TrimSpace(os.Getenv("OIDC_CLIENT_ID"))

	cfg := AuthConfig{
		enabled: true,
	}

	if len(oidcIssuer) == 0 || len(oidcClientId) == 0 {
		cfg.enabled = false
		slog.Info("OIDC authentication disabled")
		return &cfg, nil
	}
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	oidcProvider, err := oidc.NewProvider(ctxWithTimeout, oidcIssuer)
	if err != nil {
		if errors.Is(ctxWithTimeout.Err(), context.DeadlineExceeded) {
			slog.Error("Failed to initialize OIDC provider", slog.String("error", "timeout"), slog.String("issuer", oidcIssuer))
		} else {
			slog.Error("Failed to initialize OIDC provider", slog.String("error", err.Error()), slog.String("issuer", oidcIssuer))
		}
		return nil, err
	}
	slog.Info("OIDC authentication enabled",
		slog.String("issuer", oidcIssuer),
		slog.String("client_id", oidcClientId))
	cfg.Verifier = oidcProvider.Verifier(&oidc.Config{ClientID: oidcClientId})

	return &cfg, nil
}
