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
		wantErr      bool
	}{
		{
			name: "All variables unset",
			env: map[string]string{
				"OIDC_ISSUER":    "",
				"OIDC_CLIENT_ID": "",
			},
			wantEnabled: false,
			wantErr:     false,
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
				"OIDC_CLIENT_ID": "audience",
			},
			wantErr: true,
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
				"OIDC_CLIENT_ID": "audience",
			},
			wantErr: true,
		},
		{
			name: "Whitespace variables",
			env: map[string]string{
				"OIDC_ISSUER":    "   ",
				"OIDC_AUDI	ENCE": "   ",
			},
			wantEnabled: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars first to ensure clean state
			t.Setenv("OIDC_ISSUER", "")
			t.Setenv("OIDC_CLIENT_ID", "")

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

			if cfg.Enabled() != tt.wantEnabled {
				t.Errorf("NewAuthConfig() Enabled = %v, want %v", cfg.Enabled(), tt.wantEnabled)
			}
			if cfg.Issuer() != tt.wantIssuer {
				t.Errorf("NewAuthConfig() Issuer = %v, want %v", cfg.Issuer(), tt.wantIssuer)
			}
			if cfg.ClientId() != tt.wantAudience {
				t.Errorf("NewAuthConfig() Audience = %v, want %v", cfg.ClientId(), tt.wantAudience)
			}
		})
	}
}
