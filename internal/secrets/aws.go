package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type AWSProvider struct {
	client *secretsmanager.Client
}

func (p *AWSProvider) GetSecret(ctx context.Context, name string) (string, error) {
	out, err := p.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if err != nil {
		return "", fmt.Errorf("fetching secret %q from AWS: %w", name, err)
	}
	if out.SecretString == nil {
		return "", fmt.Errorf("secret %q has no string value", name)
	}
	return *out.SecretString, nil
}

func (p *AWSProvider) GetDatabaseInfo(ctx context.Context) (DatabaseInfo, error) {
	secretName, found := os.LookupEnv("DATABASE_SECRET_NAME")
	if !found {
		return DatabaseInfo{}, fmt.Errorf("DATABASE_SECRET_NAME not set")
	}
	jsonSecret, err := p.GetSecret(ctx, secretName)
	if err != nil {
		return DatabaseInfo{}, err
	}
	secret := DatabaseInfo{}
	err = json.Unmarshal([]byte(jsonSecret), &secret)
	if err != nil {
		return DatabaseInfo{}, err
	}
	return secret, nil
}
