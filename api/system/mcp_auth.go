package system

import (
	"encoding/json"
	"net/http"
	"service-atlas/internal/auth"
)

type MCPAuthMode string

const (
	MCPAuthModeEnabled  MCPAuthMode = "enabled"
	MCPAuthModeDisabled MCPAuthMode = "disabled"
	MCPAuthModeNone     MCPAuthMode = "none"
)

type MCPAuthConfigResponse struct {
	AuthMode MCPAuthMode  `json:"auth_mode"`
	OIDC     *MCPOIDCInfo `json:"oidc,omitempty"`
	Reason   string       `json:"reason,omitempty"`
}

type MCPOIDCInfo struct {
	Issuer   string `json:"issuer"`
	ClientID string `json:"client_id"`
	Audience string `json:"audience"`
}

func CreateMCPAuthEndpoint(cfg *auth.Config) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		resp := MCPAuthConfigResponse{}
		if cfg.Enabled() {
			resp.AuthMode = MCPAuthModeEnabled
			resp.OIDC = &MCPOIDCInfo{
				Issuer:   cfg.Issuer(),
				ClientID: cfg.MCPClientId(),
				Audience: cfg.Audience(),
			}
			if resp.OIDC.ClientID == "" {
				resp.AuthMode = MCPAuthModeNone
				resp.Reason = "OIDC client ID not set"
				resp.OIDC = nil
			}
		} else {
			resp.AuthMode = MCPAuthModeDisabled
		}

		encoder := json.NewEncoder(w)

		if err := encoder.Encode(resp); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
