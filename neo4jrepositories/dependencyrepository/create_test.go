package dependencyrepository

import (
	"context"
	"errors"
	"service-atlas/internal/customerrors"
	"service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jDependencyRepository_AddDependency_WithVersion(t *testing.T) {
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

	// Arrange: create two services
	serviceID := "11111111-1111-1111-1111-111111111111"
	depID := "22222222-2222-2222-2222-222222222222"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx,
		"CREATE (s1:Service {id: $sid, name: $sname}) RETURN s1",
		map[string]any{"sid": serviceID, "sname": "svc-a"},
	); err != nil {
		t.Fatalf("failed to create service1: %v", err)
	}
	if _, err = write.Run(ctx,
		"CREATE (s2:Service {id: $did, name: $dname}) RETURN s2",
		map[string]any{"did": depID, "dname": "svc-b"},
	); err != nil {
		t.Fatalf("failed to create service2: %v", err)
	}

	// Act
	dep := repositories.Dependency{Id: depID, Version: "1.2.3"}
	if err := repo.AddDependency(ctx, serviceID, dep); err != nil {
		t.Fatalf("AddDependency returned error: %v", err)
	}

	// Assert: relationship exists with version and interaction_type
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[r:DEPENDS_ON]->(:Service {id: $did}) RETURN r.version AS version, r.interaction_type AS it",
		map[string]any{"sid": serviceID, "did": depID},
	)
	if err != nil {
		t.Fatalf("failed to verify dependency: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying dependency: %v", err)
	}
	ver, ok := rec.Get("version")
	if !ok {
		t.Fatalf("missing version property")
	}
	if ver != "1.2.3" {
		t.Fatalf("expected version %q, got %#v", "1.2.3", ver)
	}
	it, ok := rec.Get("it")
	if !ok {
		t.Fatalf("missing interaction_type property")
	}
	if it != "data" {
		t.Fatalf("expected interaction_type %q, got %#v", "data", it)
	}
}

func TestNeo4jDependencyRepository_AddDependency_WithInteractionType(t *testing.T) {
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

	// Arrange: create two services
	serviceID := "11111111-1111-1111-1111-111111111111"
	depID := "22222222-2222-2222-2222-222222222222"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx,
		"CREATE (s1:Service {id: $sid, name: $sname}) RETURN s1",
		map[string]any{"sid": serviceID, "sname": "svc-a"},
	); err != nil {
		t.Fatalf("failed to create service1: %v", err)
	}
	if _, err = write.Run(ctx,
		"CREATE (s2:Service {id: $did, name: $dname}) RETURN s2",
		map[string]any{"did": depID, "dname": "svc-b"},
	); err != nil {
		t.Fatalf("failed to create service2: %v", err)
	}

	// Act
	dep := repositories.Dependency{Id: depID, InteractionType: "async"}
	if err := repo.AddDependency(ctx, serviceID, dep); err != nil {
		t.Fatalf("AddDependency returned error: %v", err)
	}

	// Assert: relationship exists with interaction_type
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[r:DEPENDS_ON]->(:Service {id: $did}) RETURN r.interaction_type AS it",
		map[string]any{"sid": serviceID, "did": depID},
	)
	if err != nil {
		t.Fatalf("failed to verify dependency: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying dependency: %v", err)
	}
	it, ok := rec.Get("it")
	if !ok {
		t.Fatalf("missing interaction_type property")
	}
	if it != "async" {
		t.Fatalf("expected interaction_type %q, got %#v", "async", it)
	}
}

func TestNeo4jDependencyRepository_AddDependency_WithoutVersion(t *testing.T) {
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

	// Arrange
	sid := "33333333-3333-3333-3333-333333333333"
	did := "44444444-4444-4444-4444-444444444444"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-c'}) RETURN s", map[string]any{"id": sid}); err != nil {
		t.Fatalf("create sid: %v", err)
	}
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-d'}) RETURN s", map[string]any{"id": did}); err != nil {
		t.Fatalf("create did: %v", err)
	}

	// Act
	dep := repositories.Dependency{Id: did}
	if err := repo.AddDependency(ctx, sid, dep); err != nil {
		t.Fatalf("AddDependency error: %v", err)
	}

	// Assert: relationship exists without version property
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[r:DEPENDS_ON]->(:Service {id: $did}) RETURN r.version AS version",
		map[string]any{"sid": sid, "did": did},
	)
	if err != nil {
		t.Fatalf("failed to verify dependency: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record: %v", err)
	}
	ver, ok := rec.Get("version")
	if !ok {
		t.Fatalf("missing version property")
	}
	if ver != nil && ver != "" {
		t.Fatalf("expected no version property, got %#v", ver)
	}
}

func TestNeo4jDependencyRepository_AddDependency_NotFound(t *testing.T) {
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

	// Missing both services
	err = repo.AddDependency(ctx, "00000000-0000-0000-0000-000000000000", repositories.Dependency{Id: "99999999-9999-9999-9999-999999999999"})
	if err == nil {
		t.Fatalf("expected error when services not found")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d (msg=%q)", httpErr.Status, httpErr.Msg)
	}
}
