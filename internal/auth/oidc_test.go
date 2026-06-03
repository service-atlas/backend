package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startMockOIDCServer() *httptest.Server {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	// OIDC Discovery
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]any{
			"issuer":                                server.URL,
			"jwks_uri":                              fmt.Sprintf("%s/keys", server.URL),
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	})

	// JWKS
	mux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		jwks := map[string]any{
			"keys": []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	return server
}

func TestNewAuthConfig(t *testing.T) {
	server := startMockOIDCServer()
	defer server.Close()

	tests := []struct {
		name            string
		env             map[string]string
		wantEnabled     bool
		wantIssuer      string
		wantAudience    string
		wantMCPClientID string
	}{
		{
			name: "All variables unset",
			env: map[string]string{
				"OIDC_ISSUER":        "",
				"OIDC_MCP_CLIENT_ID": "",
				"OIDC_AUDIENCE":      "",
			},
			wantEnabled: false,
		},
		{
			name: "All variables set",
			env: map[string]string{
				"OIDC_ISSUER":        server.URL,
				"OIDC_MCP_CLIENT_ID": "client-id",
				"OIDC_AUDIENCE":      "audience",
			},
			wantEnabled:     true,
			wantIssuer:      server.URL,
			wantAudience:    "audience",
			wantMCPClientID: "client-id",
		},
		{
			name: "Only Issuer set",
			env: map[string]string{
				"OIDC_ISSUER": server.URL,
			},
			wantEnabled: false,
		},
		{
			name: "Only MCP Client ID set",
			env: map[string]string{
				"OIDC_MCP_CLIENT_ID": "client-id",
			},
			wantEnabled: false,
		},
		{
			name: "Whitespace variables",
			env: map[string]string{
				"OIDC_ISSUER":        "   ",
				"OIDC_MCP_CLIENT_ID": "   ",
				"OIDC_AUDIENCE":      "   ",
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars first to ensure clean state
			t.Setenv("OIDC_ISSUER", "")
			t.Setenv("OIDC_MCP_CLIENT_ID", "")
			t.Setenv("OIDC_AUDIENCE", "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := NewAuthConfig()

			if err != nil && tt.wantEnabled {
				t.Fatalf("NewAuthConfig() error = %v", err)
			}

			if cfg.Enabled() != tt.wantEnabled {
				t.Errorf("NewAuthConfig() Enabled = %v, want %v", cfg.Enabled(), tt.wantEnabled)
			}

			if tt.wantEnabled {
				if cfg.Issuer() != tt.wantIssuer {
					t.Errorf("NewAuthConfig() Issuer = %v, want %v", cfg.Issuer(), tt.wantIssuer)
				}
				if cfg.Audience() != tt.wantAudience {
					t.Errorf("NewAuthConfig() Audience = %v, want %v", cfg.Audience(), tt.wantAudience)
				}
				if cfg.MCPClientID() != tt.wantMCPClientID {
					t.Errorf("NewAuthConfig() MCPClientId = %v, want %v", cfg.MCPClientID(), tt.wantMCPClientID)
				}
			}
		})
	}
}

func TestAuthConfigGetters(t *testing.T) {
	cfg := NewTestAuthConfig(true, "issuer", "audience", "client")

	if !cfg.Enabled() {
		t.Error("Enabled() should be true")
	}
	if cfg.Issuer() != "issuer" {
		t.Errorf("Issuer() = %v, want issuer", cfg.Issuer())
	}
	if cfg.Audience() != "audience" {
		t.Errorf("Audience() = %v, want audience", cfg.Audience())
	}
	if cfg.MCPClientID() != "client" {
		t.Errorf("MCPClientId() = %v, want client", cfg.MCPClientID())
	}
	if cfg.Verifier() != nil {
		t.Error("Verifier() should be nil in test config without manual set")
	}
}
