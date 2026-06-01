package auth

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
)

type TokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

type Claims struct {
	Subject string
	Name    string
	Email   string
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

func Middleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken := extractBearerToken(r.Header.Get("Authorization"))
			if rawToken == "" {
				http.Error(w, "Unauthorized: missing or invalid token", http.StatusUnauthorized)
				return
			}
			token, err := verifier.Verify(r.Context(), rawToken)
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
				nameCtx = context.WithValue(r.Context(), nameContextKey{}, claims.Name)
			} else if claims.Email != "" {
				nameCtx = context.WithValue(r.Context(), nameContextKey{}, claims.Email)
			} else if claims.Subject != "" {
				nameCtx = context.WithValue(r.Context(), nameContextKey{}, claims.Subject)
			}

			next.ServeHTTP(w, r.WithContext(nameCtx))
		})
	}
}
