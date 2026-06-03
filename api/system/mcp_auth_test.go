package system

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"service-atlas/internal/auth"
	"testing"
)

func TestCreateMCPAuthEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *auth.Config
		expectedStatus int
		expectedBody   MCPAuthConfigResponse
	}{
		{
			name:           "Auth Enabled",
			cfg:            auth.NewTestAuthConfig(true, "https://issuer", "audience", "client-id"),
			expectedStatus: http.StatusOK,
			expectedBody: MCPAuthConfigResponse{
				AuthMode: MCPAuthModeEnabled,
				OIDC: &MCPOIDCInfo{
					Issuer:   "https://issuer",
					ClientID: "client-id",
					Audience: "audience",
				},
			},
		},
		{
			name:           "Auth Disabled",
			cfg:            auth.NewTestAuthConfig(false, "", "", ""),
			expectedStatus: http.StatusOK,
			expectedBody: MCPAuthConfigResponse{
				AuthMode: MCPAuthModeDisabled,
			},
		},
		{
			name:           "Auth Enabled but missing ClientID",
			cfg:            auth.NewTestAuthConfig(true, "https://issuer", "audience", ""),
			expectedStatus: http.StatusOK,
			expectedBody: MCPAuthConfigResponse{
				AuthMode: MCPAuthModeNone,
				Reason:   "OIDC client ID not set",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CreateMCPAuthEndpoint(tt.cfg)
			req := httptest.NewRequest(http.MethodGet, "/auth/mcp/config", nil)
			rr := httptest.NewRecorder()

			handler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			var resp MCPAuthConfigResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			if resp.AuthMode != tt.expectedBody.AuthMode {
				t.Errorf("expected AuthMode %s, got %s", tt.expectedBody.AuthMode, resp.AuthMode)
			}

			if tt.expectedBody.OIDC != nil {
				if resp.OIDC == nil {
					t.Fatal("expected OIDC info, got nil")
				}
				if resp.OIDC.Issuer != tt.expectedBody.OIDC.Issuer {
					t.Errorf("expected Issuer %s, got %s", tt.expectedBody.OIDC.Issuer, resp.OIDC.Issuer)
				}
				if resp.OIDC.ClientID != tt.expectedBody.OIDC.ClientID {
					t.Errorf("expected ClientID %s, got %s", tt.expectedBody.OIDC.ClientID, resp.OIDC.ClientID)
				}
				if resp.OIDC.Audience != tt.expectedBody.OIDC.Audience {
					t.Errorf("expected Audience %s, got %s", tt.expectedBody.OIDC.Audience, resp.OIDC.Audience)
				}
			} else if resp.OIDC != nil {
				t.Errorf("expected no OIDC info, got %+v", resp.OIDC)
			}

			if resp.Reason != tt.expectedBody.Reason {
				t.Errorf("expected Reason %s, got %s", tt.expectedBody.Reason, resp.Reason)
			}
		})
	}
}
