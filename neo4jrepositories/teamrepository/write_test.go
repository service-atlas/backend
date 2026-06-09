package teamrepository

import (
	"context"
	"service-atlas/internal/auth"
	"service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jTeamRepository_CreateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()
	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tc.Container.Terminate(ctx)
	})

	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = driver.Close(ctx)
	}()
	repo := New(driver)
	team := repositories.Team{
		Name: "test",
	}
	now := time.Now()
	_, err = repo.CreateTeam(ctx, team)
	if err != nil {
		t.Fatal(err)
	}
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})

	defer func() {
		_ = session.Close(ctx)
	}()

	// Query only the specific team we just created to avoid nondeterministic results
	result, err := session.Run(ctx,
		"MATCH (n:Team {name: $name}) RETURN n.name as name, n.id as id, n.created as created",
		map[string]any{"name": team.Name},
	)
	if err != nil {
		t.Fatal(err)
	}

	returnedTeam, err := result.Single(ctx)
	if err != nil || returnedTeam == nil {
		t.Fatalf("expected single team record, got error: %v", err)
	}

	// Validate name exists, is string, and matches
	nameVal, ok := returnedTeam.Get("name")
	if !ok {
		t.Fatalf("missing 'name' field in record")
	}
	nameStr, ok := nameVal.(string)
	if !ok {
		t.Fatalf("field 'name' is not a string: %T", nameVal)
	}
	if nameStr != team.Name {
		t.Errorf("expected name %q, got %q", team.Name, nameStr)
	}

	// Validate id exists, is string, and non-empty
	idVal, ok := returnedTeam.Get("id")
	if !ok {
		t.Fatalf("missing 'id' field in record")
	}
	idStr, ok := idVal.(string)
	if !ok {
		t.Fatalf("field 'id' is not a string: %T", idVal)
	}
	if idStr == "" {
		t.Errorf("expected non-empty 'id', got empty string")
	}

	// Validate created exists and is a temporal value; accept time.Time or convert supported Neo4j temporal types
	createdVal, ok := returnedTeam.Get("created")
	if !ok {
		t.Fatalf("missing 'created' field in record")
	}

	var createdTime time.Time
	switch c := createdVal.(type) {
	case time.Time:
		createdTime = c
	default:
		// If other Neo4j temporal types appear in the future, report clearly
		t.Fatalf("unsupported 'created' type %T; expected time.Time", createdVal)
	}

	// Bounds check to avoid flaky exact-equality time comparisons
	if createdTime.Before(now) || createdTime.After(now.Add(10*time.Second)) {
		t.Errorf("expected 'created' between %s and %s, got %s", now, now.Add(10*time.Second), createdTime)
	}
}

func TestNeo4jTeamRepository_CreateTeam_WithCreatedBy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := auth.ContextWithName(context.Background(), "John Doe")
	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tc.Container.Terminate(ctx)
	})

	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = driver.Close(ctx)
	}()
	repo := New(driver)
	team := repositories.Team{
		Name: "test-created-by",
	}
	_, err = repo.CreateTeam(ctx, team)
	if err != nil {
		t.Fatal(err)
	}
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeRead,
	})

	defer func() {
		_ = session.Close(ctx)
	}()

	result, err := session.Run(ctx,
		"MATCH (n:Team {name: $name}) RETURN n.createdBy as createdBy",
		map[string]any{"name": team.Name},
	)
	if err != nil {
		t.Fatal(err)
	}

	record, err := result.Single(ctx)
	if err != nil || record == nil {
		t.Fatalf("expected single team record, got error: %v", err)
	}

	cbVal, _ := record.Get("createdBy")
	if cbVal != "John Doe" {
		t.Errorf("expected createdBy %q, got %q", "John Doe", cbVal)
	}
}

func TestUpdateTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	// Start Neo4j test container
	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tc.Container.Terminate(ctx)
	})

	// Connect driver
	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	// Create a team we can fetch
	team := repositories.Team{Name: "old-team-name"}
	now := time.Now()
	id, err := repo.CreateTeam(ctx, team)
	if err != nil {
		t.Fatal(err)
	}
	team.Id = id
	team.Name = "new-team-name"
	err = repo.UpdateTeam(ctx, team)
	if err != nil {
		t.Fatal(err)
	}

	// Read session to fetch the created team's id by name (deterministic lookup)
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = session.Close(ctx) }()

	res, err := session.Run(ctx,
		"MATCH (n:Team {name: $name}) RETURN n.name AS name, n.updated AS updated",
		map[string]any{"name": team.Name},
	)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := res.Single(ctx)
	if err != nil || rec == nil {
		t.Fatalf("expected single record, got error: %v", err)
	}

	nameVal, ok := rec.Get("name")
	if !ok {
		t.Fatalf("missing 'name' in created team record")
	}
	nameStr, ok := nameVal.(string)
	if !ok || nameStr == "" {
		t.Fatalf("expected non-empty 'name' in created team record")
	}
	if nameStr != team.Name {
		t.Fatalf("expected name %q, got %q", team.Name, nameStr)
	}
	updatedVal, ok := rec.Get("updated")
	if !ok {
		t.Fatalf("missing 'updated' in created team record")
	}
	updatedTime, ok := updatedVal.(time.Time)
	if !ok || updatedTime.IsZero() {
		t.Fatalf("expected non-zero 'updated' in created team record")
	}
	if updatedTime.Before(now) || updatedTime.After(now.Add(10*time.Second)) {
		t.Fatalf("expected 'updated' between %s and %s, got %s", now, now.Add(10*time.Second), updatedTime)
	}

}

func TestUpdateTeam_WithUpdatedBy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	// Start Neo4j test container
	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tc.Container.Terminate(ctx)
	})

	// Connect driver
	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	// Create a team
	team := repositories.Team{Name: "old-team"}
	id, err := repo.CreateTeam(ctx, team)
	if err != nil {
		t.Fatal(err)
	}

	// Update with updatedBy
	updateCtx := auth.ContextWithName(ctx, "Jane Doe")
	team.Id = id
	team.Name = "updated-team"
	err = repo.UpdateTeam(updateCtx, team)
	if err != nil {
		t.Fatal(err)
	}

	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = session.Close(ctx) }()

	res, err := session.Run(ctx,
		"MATCH (n:Team {id: $id}) RETURN n.updatedBy AS updatedBy",
		map[string]any{"id": id},
	)
	if err != nil {
		t.Fatal(err)
	}
	rec, err := res.Single(ctx)
	if err != nil || rec == nil {
		t.Fatalf("expected single record, got error: %v", err)
	}

	ubVal, _ := rec.Get("updatedBy")
	if ubVal != "Jane Doe" {
		t.Fatalf("expected updatedBy %q, got %q", "Jane Doe", ubVal)
	}
}
