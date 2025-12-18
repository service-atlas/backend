package dependencyrepository

import (
	"context"
	"errors"
	"service-atlas/internal/customerrors"
	"service-atlas/neo4jrepositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jDependencyRepository_DeleteDependency_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	// Start Neo4j test container
	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(ctx) })

	// Connect driver
	driver, err := neo4j.NewDriverWithContext(tc.Endpoint, neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	// Arrange: create two services and a dependency
	sid := "12121212-1212-1212-1212-121212121212"
	did := "34343434-3434-3434-3434-343434343434"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-x'}) RETURN s", map[string]any{"id": sid}); err != nil {
		t.Fatalf("create sid: %v", err)
	}
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-y'}) RETURN s", map[string]any{"id": did}); err != nil {
		t.Fatalf("create did: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON {version: '9.9.9'}]->(b)", map[string]any{"a": sid, "b": did}); err != nil {
		t.Fatalf("rel a->b: %v", err)
	}

	// Act
	if err := repo.DeleteDependency(ctx, sid, did); err != nil {
		t.Fatalf("DeleteDependency returned error: %v", err)
	}

	// Assert: relationship is gone
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[:DEPENDS_ON]->(:Service {id: $did}) RETURN count(*) as cnt",
		map[string]any{"sid": sid, "did": did},
	)
	if err != nil {
		t.Fatalf("verify delete: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record: %v", err)
	}
	cnt, ok := rec.Get("cnt")
	if !ok {
		t.Fatalf("missing 'cnt' property")
	}
	if cnt.(int64) != 0 {
		t.Fatalf("expected no relationship, found %d", cnt.(int64))
	}
}

func TestNeo4jDependencyRepository_DeleteDependency_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
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

	// No relationship exists (and services likely don't either)
	err = repo.DeleteDependency(ctx, "00000000-0000-0000-0000-000000000000", "ffffffff-ffff-ffff-ffff-ffffffffffff")
	if err == nil {
		t.Fatalf("expected error when dependency relationship not found")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.Status)
	}
}
