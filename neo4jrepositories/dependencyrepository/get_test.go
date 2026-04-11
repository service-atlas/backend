package dependencyrepository

import (
	"context"
	"errors"
	"service-atlas/internal/customerrors"
	"service-atlas/neo4jrepositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jDependencyRepository_GetDependencies_Success(t *testing.T) {
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

	// Arrange: s1 depends on s2 and s3
	s1 := "55555555-5555-5555-5555-555555555555"
	s2 := "66666666-6666-6666-6666-666666666666"
	s3 := "77777777-7777-7777-7777-777777777777"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-1', type: 'api'}) RETURN s", map[string]any{"id": s1}); err != nil {
		t.Fatalf("create s1: %v", err)
	}
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-2', type: 'db'}) RETURN s", map[string]any{"id": s2}); err != nil {
		t.Fatalf("create s2: %v", err)
	}
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-3', type: 'queue'}) RETURN s", map[string]any{"id": s3}); err != nil {
		t.Fatalf("create s3: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON {version: '2.0.0', interaction_type: 'security'}]->(b)", map[string]any{"a": s1, "b": s2}); err != nil {
		t.Fatalf("rel s1->s2: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": s1, "b": s3}); err != nil {
		t.Fatalf("rel s1->s3: %v", err)
	}

	// Act
	deps, err := repo.GetDependencies(ctx, s1)
	if err != nil {
		t.Fatalf("GetDependencies returned error: %v", err)
	}

	// Assert
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(deps))
	}
	// verify that both ids are present and versions mapped correctly (order is not guaranteed)
	found := map[string]bool{"svc-2": false, "svc-3": false}
	for _, d := range deps {
		if d.Id == s2 {
			found["svc-2"] = true
			if d.Version != "2.0.0" {
				t.Fatalf("expected version 2.0.0 for %s, got %q", s2, d.Version)
			}
			if d.InteractionType != "security" {
				t.Fatalf("expected interaction_type 'security' for %s, got %q", s2, d.InteractionType)
			}
			if d.Name == "" || d.ServiceType == "" {
				t.Fatalf("expected name and type for %s", s2)
			}
		}
		if d.Id == s3 {
			found["svc-3"] = true
			if d.Version != "" {
				t.Fatalf("expected empty version for %s, got %q", s3, d.Version)
			}
			if d.InteractionType != "data" {
				t.Fatalf("expected interaction_type 'data' (default) for %s, got %q", s3, d.InteractionType)
			}
			if d.Name == "" || d.ServiceType == "" {
				t.Fatalf("expected name and type for %s", s3)
			}
		}
	}
	if !found["svc-2"] || !found["svc-3"] {
		t.Fatalf("missing expected dependencies: %+v", found)
	}
}

func TestNeo4jDependencyRepository_GetDependencies_NotFound(t *testing.T) {
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

	_, err = repo.GetDependencies(ctx, "00000000-0000-0000-0000-000000000000")
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

func TestNeo4jDependencyRepository_GetDependents_Success(t *testing.T) {
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

	// Arrange: sA and sC depend on sB
	sA := "88888888-8888-8888-8888-888888888888"
	sB := "99999999-9999-9999-9999-999999999999"
	sC := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-A', type: 'api'}) RETURN s", map[string]any{"id": sA}); err != nil {
		t.Fatalf("create sA: %v", err)
	}
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-B', type: 'db'}) RETURN s", map[string]any{"id": sB}); err != nil {
		t.Fatalf("create sB: %v", err)
	}
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-C', type: 'worker'}) RETURN s", map[string]any{"id": sC}); err != nil {
		t.Fatalf("create sC: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON {version: '0.1.0', interaction_type: 'async'}]->(b)", map[string]any{"a": sA, "b": sB}); err != nil {
		t.Fatalf("rel sA->sB: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": sC, "b": sB}); err != nil {
		t.Fatalf("rel sC->sB: %v", err)
	}

	// Act
	deps, err := repo.GetDependents(ctx, sB)
	if err != nil {
		t.Fatalf("GetDependents returned error: %v", err)
	}

	// Assert: expect 2 dependents A and C
	if len(deps) != 2 {
		t.Fatalf("expected 2 dependents, got %d", len(deps))
	}
	seen := map[string]bool{"A": false, "C": false}
	for _, d := range deps {
		if d.Id == sA {
			seen["A"] = true
			if d.Version != "0.1.0" {
				t.Fatalf("expected version for A->B to be 0.1.0, got %q", d.Version)
			}
			if d.InteractionType != "async" {
				t.Fatalf("expected interaction_type 'async' for A->B, got %q", d.InteractionType)
			}
			if d.Name == "" || d.ServiceType == "" {
				t.Fatalf("expected name and type for A")
			}
		}
		if d.Id == sC {
			seen["C"] = true
			if d.Version != "" {
				t.Fatalf("expected empty version for C->B, got %q", d.Version)
			}
			if d.InteractionType != "data" {
				t.Fatalf("expected interaction_type 'data' (default) for C->B, got %q", d.InteractionType)
			}
			if d.Name == "" || d.ServiceType == "" {
				t.Fatalf("expected name and type for C")
			}
		}
	}
	if !seen["A"] || !seen["C"] {
		t.Fatalf("missing expected dependents: %+v", seen)
	}
}

func TestNeo4jDependencyRepository_GetDependents_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cErr := tc.Container.Terminate(context.Background())
		if cErr != nil {
			t.Logf("error terminating container: %v", cErr)
		}
	})

	driver, err := neo4j.NewDriverWithContext(tc.Endpoint, neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	_, err = repo.GetDependents(ctx, "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
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
