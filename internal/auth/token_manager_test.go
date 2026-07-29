package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/service-atlas/secrets-provider"
)

// mockProvider implements secretsprovider.Provider interface
type mockProvider struct {
	dbInfo secretsprovider.DatabaseInfo
	err    error
}

func (m *mockProvider) GetDatabaseInfo(_ context.Context) (secretsprovider.DatabaseInfo, error) {
	return m.dbInfo, m.err
}

func (m *mockProvider) GetSecret(_ context.Context, _ string) (string, error) {
	return "", nil
}

func TestSecretProviderTokenManager_GetAuthToken(t *testing.T) {
	ctx := t.Context()

	t.Run("success", func(t *testing.T) {
		expectedInfo := secretsprovider.DatabaseInfo{
			Username: "neo4j",
			Password: "password",
			URL:      "bolt://localhost:7687",
		}
		mock := &mockProvider{dbInfo: expectedInfo}
		tm := NewSecretProviderTokenManager(mock)

		token, err := tm.GetAuthToken(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// BasicAuth returns a map[string]any under the hood
		authMap := token.Tokens
		if authMap["scheme"] != "basic" {
			t.Errorf("expected scheme basic, got %v", authMap["scheme"])
		}
		if authMap["principal"] != expectedInfo.Username {
			t.Errorf("expected principal %s, got %v", expectedInfo.Username, authMap["principal"])
		}
		if authMap["credentials"] != expectedInfo.Password {
			t.Errorf("expected credentials %s, got %v", expectedInfo.Password, authMap["credentials"])
		}
	})

	t.Run("error from provider", func(t *testing.T) {
		expectedErr := errors.New("provider error")
		mock := &mockProvider{err: expectedErr}
		tm := NewSecretProviderTokenManager(mock)

		_, err := tm.GetAuthToken(ctx)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestSecretProviderTokenManager_HandleSecurityException(t *testing.T) {
	tm := NewSecretProviderTokenManager(&mockProvider{})
	ctx := t.Context()

	t.Run("refreshable security error", func(t *testing.T) {
		securityErr := &neo4j.Neo4jError{Code: "Neo.ClientError.Security.Unauthorized"}
		refreshed, err := tm.HandleSecurityException(ctx, neo4j.AuthToken{}, securityErr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !refreshed {
			t.Error("expected refreshed to be true for Unauthorized error")
		}
	})

	t.Run("non-refreshable security error", func(t *testing.T) {
		securityErr := &neo4j.Neo4jError{Code: "Neo.ClientError.Security.Forbidden"}
		refreshed, err := tm.HandleSecurityException(ctx, neo4j.AuthToken{}, securityErr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if refreshed {
			t.Error("expected refreshed to be false for Forbidden error")
		}
	})
}
