package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAWSProvider_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := t.Context()

	// Start Ministack container
	// Ministack is a lightweight AWS mock. It typically listens on port 4566.
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "ministackorg/ministack:latest",
			ExposedPorts: []string{"4566/tcp"},
			WaitingFor:   wait.ForListeningPort("4566/tcp"),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("failed to start ministack container: %v", err)
	}
	defer container.Terminate(ctx)

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "4566")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())

	// Configure AWS SDK to use Ministack
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "testing", SecretAccessKey: "testing", SessionToken: "testing",
			},
		}),
	)
	if err != nil {
		t.Fatalf("failed to load AWS config: %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	provider := &AWSProvider{client: client}

	t.Run("GetSecret", func(t *testing.T) {
		secretName := "my-test-secret"
		secretValue := "my-secret-value"

		// Pre-create the secret in Ministack
		_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
			Name:         aws.String(secretName),
			SecretString: aws.String(secretValue),
		})
		if err != nil {
			t.Fatalf("failed to create secret in ministack: %v", err)
		}

		val, err := provider.GetSecret(ctx, secretName)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if val != secretValue {
			t.Errorf("expected %q, got %q", secretValue, val)
		}
	})

	t.Run("GetDatabaseInfo", func(t *testing.T) {
		secretName := "db-secret"
		dbInfo := DatabaseInfo{
			URL:      "bolt://neo4j:7687",
			Username: "neo4j",
			Password: "password123",
		}
		jsonBytes, _ := json.Marshal(dbInfo)

		// Pre-create the secret
		_, err := client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
			Name:         aws.String(secretName),
			SecretString: aws.String(string(jsonBytes)),
		})
		if err != nil {
			t.Fatalf("failed to create db secret in ministack: %v", err)
		}

		t.Setenv("DATABASE_SECRET_NAME", secretName)

		info, err := provider.GetDatabaseInfo(ctx)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if info != dbInfo {
			t.Errorf("expected %+v, got %+v", dbInfo, info)
		}
	})

	t.Run("GetDatabaseInfo_EnvMissing", func(t *testing.T) {
		os.Unsetenv("DATABASE_SECRET_NAME")
		_, err := provider.GetDatabaseInfo(ctx)
		if err == nil {
			t.Error("expected error when DATABASE_SECRET_NAME is not set, got nil")
		}
	})

	t.Run("GetSecret_NotFound", func(t *testing.T) {
		_, err := provider.GetSecret(ctx, "non-existent")
		if err == nil {
			t.Error("expected error for non-existent secret, got nil")
		}
	})
}
