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
		wantJWKSURL  string
		wantErr      bool
	}{
		{
			name: "All variables unset",
			env: map[string]string{
				"OIDC_ISSUER":   "",
				"OIDC_AUDIENCE": "",
				"OIDC_JWKS_URL": "",
			},
			wantEnabled: false,
			wantErr:     false,
		},
		{
			name: "All variables set",
			env: map[string]string{
				"OIDC_ISSUER":   "https://issuer.com",
				"OIDC_AUDIENCE": "audience",
				"OIDC_JWKS_URL": "https://issuer.com/.well-known/jwks.json",
			},
			wantEnabled:  true,
			wantIssuer:   "https://issuer.com",
			wantAudience: "audience",
			wantJWKSURL:  "https://issuer.com/.well-known/jwks.json",
			wantErr:      false,
		},
		{
			name: "Only Issuer set",
			env: map[string]string{
				"OIDC_ISSUER": "https://issuer.com",
			},
			wantErr: true,
		},
		{
			name: "Only Audience set",
			env: map[string]string{
				"OIDC_AUDIENCE": "audience",
			},
			wantErr: true,
		},
		{
			name: "Only JWKS URL set",
			env: map[string]string{
				"OIDC_JWKS_URL": "https://issuer.com/.well-known/jwks.json",
			},
			wantErr: true,
		},
		{
			name: "Issuer and Audience set, JWKS URL missing",
			env: map[string]string{
				"OIDC_ISSUER":   "https://issuer.com",
				"OIDC_AUDIENCE": "audience",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars first to ensure clean state
			t.Setenv("OIDC_ISSUER", "")
			t.Setenv("OIDC_AUDIENCE", "")
			t.Setenv("OIDC_JWKS_URL", "")

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := NewAuthConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAuthConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if cfg.Enabled != tt.wantEnabled {
				t.Errorf("NewAuthConfig() Enabled = %v, want %v", cfg.Enabled, tt.wantEnabled)
			}
			if cfg.Issuer != tt.wantIssuer {
				t.Errorf("NewAuthConfig() Issuer = %v, want %v", cfg.Issuer, tt.wantIssuer)
			}
			if cfg.Audience != tt.wantAudience {
				t.Errorf("NewAuthConfig() Audience = %v, want %v", cfg.Audience, tt.wantAudience)
			}
			if cfg.JWKSURL != tt.wantJWKSURL {
				t.Errorf("NewAuthConfig() JWKSURL = %v, want %v", cfg.JWKSURL, tt.wantJWKSURL)
			}
		})
	}
}
