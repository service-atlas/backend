package auth

import (
	"testing"
)

func TestNewAuthConfig(t *testing.T) {
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
				"OIDC_ISSUER":        "https://issuer.com",
				"OIDC_MCP_CLIENT_ID": "client-id",
				"OIDC_AUDIENCE":      "audience",
			},
			wantEnabled:     true,
			wantIssuer:      "https://issuer.com",
			wantAudience:    "audience",
			wantMCPClientID: "client-id",
		},
		{
			name: "Only Issuer set",
			env: map[string]string{
				"OIDC_ISSUER": "https://issuer.com",
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
			},
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() && tt.wantEnabled {
				t.Skip("skipping test that requires network in short mode")
			}
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
				if cfg.MCPClientId() != tt.wantMCPClientID {
					t.Errorf("NewAuthConfig() MCPClientId = %v, want %v", cfg.MCPClientId(), tt.wantMCPClientID)
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
	if cfg.MCPClientId() != "client" {
		t.Errorf("MCPClientId() = %v, want client", cfg.MCPClientId())
	}
	if cfg.Verifier() != nil {
		t.Error("Verifier() should be nil in test config without manual set")
	}
}
