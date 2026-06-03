package auth

import (
	"testing"
)

func TestNewAuthConfig(t *testing.T) {
	tests := []struct {
		name         string
		env          map[string]string
		wantEnabled  bool
		wantIssuer   string
		wantAudience string
	}{
		{
			name: "All variables unset",
			env: map[string]string{
				"OIDC_ISSUER":    "",
				"OIDC_CLIENT_ID": "",
			},
			wantEnabled: false,
		},
		{
			name: "All variables set",
			env: map[string]string{
				"OIDC_ISSUER":    "https://issuer.com",
				"OIDC_CLIENT_ID": "audience",
			},
			wantEnabled:  true,
			wantIssuer:   "https://issuer.com",
			wantAudience: "audience",
		},
		{
			name: "Only Issuer set",
			env: map[string]string{
				"OIDC_ISSUER": "https://issuer.com",
			},
			wantEnabled: false,
		},
		{
			name: "Only Audience set",
			env: map[string]string{
				"OIDC_CLIENT_ID": "audience",
			},
			wantEnabled: false,
		},
		{
			name: "Only Issuer set",
			env: map[string]string{
				"OIDC_ISSUER": "https://issuer.com",
			},
			wantEnabled: false,
		},
		{
			name: "Only Audience set",
			env: map[string]string{
				"OIDC_CLIENT_ID": "audience",
			},
			wantEnabled: false,
		},
		{
			name: "Whitespace variables",
			env: map[string]string{
				"OIDC_ISSUER":   "   ",
				"OIDC_AUDIENCE": "   ",
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
			t.Setenv("OIDC_CLIENT_ID", "")

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
		})
	}
}
