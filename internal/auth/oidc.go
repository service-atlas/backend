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

type Config struct {
	enabled     bool
	verifier    TokenVerifier
	issuer      string
	audience    string
	mcpClientID string
}

func (cfg *Config) Verifier() TokenVerifier {
	return cfg.verifier
}
func (cfg *Config) Enabled() bool {
	return cfg.enabled
}
func (cfg *Config) MCPClientID() string {
	return cfg.mcpClientID
}
func (cfg *Config) Audience() string {
	return cfg.audience
}
func (cfg *Config) Issuer() string {
	return cfg.issuer
}

func NewTestAuthConfig(enabled bool, issuer, audience, mcpClientID string) *Config {
	return &Config{
		enabled:     enabled,
		issuer:      issuer,
		audience:    audience,
		mcpClientID: mcpClientID,
	}
}

func NewAuthConfig() (*Config, error) {
	oidcIssuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	oidcClientId := strings.TrimSpace(os.Getenv("OIDC_MCP_CLIENT_ID"))
	oidcAudience := strings.TrimSpace(os.Getenv("OIDC_AUDIENCE"))

	cfg := Config{
		enabled:     true,
		issuer:      oidcIssuer,
		audience:    oidcAudience,
		mcpClientID: oidcClientId,
	}

	if len(oidcIssuer) == 0 || len(oidcAudience) == 0 {
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
		slog.String("audience", oidcAudience))
	cfg.verifier = oidcProvider.Verifier(&oidc.Config{ClientID: oidcAudience})

	return &cfg, nil
}
