package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

type TokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type Claims struct {
	Name  string `json:"service-atlas:name,omitempty"`
	Email string `json:"service-atlas:email,omitempty"`
}

type nameContextKey struct{}

func extractBearerToken(authHeader string) string {
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return ""
	}
	return authHeader[7:]
}

func NameFromContext(ctx context.Context) string {
	if name, ok := ctx.Value(nameContextKey{}).(string); ok {
		return name
	}
	return ""
}

func ContextWithName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, nameContextKey{}, name)
}

func Middleware(authCfg *Config) func(http.Handler) http.Handler {
	if !authCfg.Enabled() || authCfg.Verifier() == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	slog.Info("Initialized OIDC Middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken := extractBearerToken(r.Header.Get("Authorization"))
			if rawToken == "" {
				http.Error(w, "Unauthorized: missing or invalid token", http.StatusUnauthorized)
				return
			}
			token, err := authCfg.Verifier().Verify(r.Context(), rawToken)
			if err != nil {
				http.Error(w, "Failed to verify token", http.StatusUnauthorized)
				return
			}
			claims := Claims{}
			if err := token.Claims(&claims); err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			nameCtx := r.Context()
			if claims.Name != "" {
				nameCtx = ContextWithName(r.Context(), claims.Name)
			} else if claims.Email != "" {
				nameCtx = ContextWithName(r.Context(), claims.Email)
			} else if token.Subject != "" {
				nameCtx = ContextWithName(r.Context(), token.Subject)
			}

			next.ServeHTTP(w, r.WithContext(nameCtx))
		})
	}
}
