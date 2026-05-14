package secrets

import (
	"testing"
)

func TestEnvProvider_GetSecret(t *testing.T) {
	p := &EnvProvider{}
	ctx := t.Context()

	t.Run("exists", func(t *testing.T) {
		t.Setenv("TEST_SECRET", "value")
		val, err := p.GetSecret(ctx, "TEST_SECRET")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if val != "value" {
			t.Errorf("expected value, got %q", val)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := p.GetSecret(ctx, "NON_EXISTENT_SECRET")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestEnvProvider_GetDatabaseInfo(t *testing.T) {
	p := &EnvProvider{}
	ctx := t.Context()

	t.Run("success", func(t *testing.T) {
		t.Setenv("DB_URL", "bolt://localhost:7687")
		t.Setenv("DB_USERNAME", "neo4j")
		t.Setenv("DB_PASSWORD", "password")

		info, err := p.GetDatabaseInfo(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if info.URL != "bolt://localhost:7687" {
			t.Errorf("expected URL bolt://localhost:7687, got %q", info.URL)
		}
		if info.Username != "neo4j" {
			t.Errorf("expected Username neo4j, got %q", info.Username)
		}
		if info.Password != "password" {
			t.Errorf("expected Password password, got %q", info.Password)
		}
	})

	t.Run("missing DB_URL", func(t *testing.T) {
		t.Setenv("DB_USERNAME", "neo4j")
		t.Setenv("DB_PASSWORD", "password")

		_, err := p.GetDatabaseInfo(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing DB_USERNAME", func(t *testing.T) {
		t.Setenv("DB_URL", "bolt://localhost:7687")
		t.Setenv("DB_PASSWORD", "password")

		_, err := p.GetDatabaseInfo(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("missing DB_PASSWORD", func(t *testing.T) {
		t.Setenv("DB_URL", "bolt://localhost:7687")
		t.Setenv("DB_USERNAME", "neo4j")

		_, err := p.GetDatabaseInfo(ctx)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
