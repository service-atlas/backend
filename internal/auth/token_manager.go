package auth

import (
	"context"
	"log/slog"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/service-atlas/secrets-provider"
)

type SecretProviderTokenManager struct {
	sProvider secretsprovider.Provider
}

func NewSecretProviderTokenManager(sProvider secretsprovider.Provider) *SecretProviderTokenManager {
	return &SecretProviderTokenManager{sProvider: sProvider}
}

func (m *SecretProviderTokenManager) GetAuthToken(ctx context.Context) (neo4j.AuthToken, error) {
	dbInfo, err := m.sProvider.GetDatabaseInfo(ctx)
	if err != nil {
		return neo4j.AuthToken{}, err
	}
	slog.Info("refreshing credentials from provider")
	return neo4j.BasicAuth(dbInfo.Username, dbInfo.Password, ""), nil
}

func (m *SecretProviderTokenManager) HandleSecurityException(_ context.Context, _ neo4j.AuthToken, _ *neo4j.Neo4jError) (bool, error) {
	return false, nil
}
