package secrets

import (
	"context"
	"fmt"
	"os"
)

type EnvProvider struct{}

func (p *EnvProvider) GetSecret(_ context.Context, name string) (string, error) {
	if val, ok := os.LookupEnv(name); ok {
		return val, nil
	}
	return "", fmt.Errorf("secret %q not found in environment", name)
}

func (p *EnvProvider) GetDatabaseInfo(ctx context.Context) (DatabaseInfo, error) {
	url, err := p.GetSecret(ctx, "DB_URL")
	if err != nil {
		return DatabaseInfo{}, err
	}
	username, err := p.GetSecret(ctx, "DB_USERNAME")
	if err != nil {
		return DatabaseInfo{}, err
	}
	password, err := p.GetSecret(ctx, "DB_PASSWORD")
	if err != nil {
		return DatabaseInfo{}, err
	}
	return DatabaseInfo{URL: url, Username: username, Password: password}, nil
}
