package secrets

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type DatabaseInfo struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Provider interface {
	GetSecret(ctx context.Context, name string) (string, error)
	GetDatabaseInfo(ctx context.Context) (DatabaseInfo, error)
}

func NewProvider(ctx context.Context) (Provider, error) {
	switch os.Getenv("SECRETS_PROVIDER") {
	case "aws":
		cfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("loading AWS config: %w", err)
		}
		return &AWSProvider{client: secretsmanager.NewFromConfig(cfg)}, nil
	default:
		return &EnvProvider{}, nil
	}
}
