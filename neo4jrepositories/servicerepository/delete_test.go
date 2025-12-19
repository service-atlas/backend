package servicerepository

import (
	"context"
	"errors"
	"service-atlas/internal/customerrors"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jServiceRepository_DeleteService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	tc, err := nRepo.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(ctx) })

	driver, err := neo4j.NewDriverWithContext(tc.Endpoint, neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	// Arrange: create a service
	createdID, err := repo.CreateService(ctx, repositories.Service{
		Name:        "svc-del",
		Description: "to delete",
		ServiceType: "api",
		Url:         "https://svc-del",
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Act
	if err := repo.DeleteService(ctx, createdID); err != nil {
		t.Fatalf("DeleteService returned error: %v", err)
	}

	// Assert: ensure node no longer exists
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx, "MATCH (s:Service {id: $id}) RETURN count(s) AS cnt", map[string]any{"id": createdID})
	if err != nil {
		t.Fatalf("failed to verify delete: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record: %v", err)
	}
	cnt, _ := rec.Get("cnt")
	if cnt.(int64) != 0 {
		t.Fatalf("expected node deleted, found count=%d", cnt.(int64))
	}
}

func TestNeo4jServiceRepository_DeleteService_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	tc, err := nRepo.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(ctx) })

	driver, err := neo4j.NewDriverWithContext(tc.Endpoint, neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	err = repo.DeleteService(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatalf("expected error when service not found")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.Status)
	}
}
