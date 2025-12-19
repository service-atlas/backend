package releaserepository

import (
	"context"
	"errors"
	"testing"
	"time"

	"service-atlas/internal/customerrors"
	"service-atlas/neo4jrepositories"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jReleaseRepository_CreateRelease_Success(t *testing.T) {
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

	// Arrange: create a Service node
	serviceID := "11111111-2222-3333-4444-555555555555"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	result, err := write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc-release', type: 'api'}) RETURN s", map[string]any{"id": serviceID})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}
	if _, err = result.Consume(ctx); err != nil {
		t.Fatalf("failed to consume result: %v", err)
	}

	// Act: create a Release with URL and Version
	relTime := time.Date(2024, 12, 25, 10, 9, 8, 0, time.UTC)
	rel := repositories.Release{
		ServiceId:   serviceID,
		ReleaseDate: relTime,
		Url:         "https://release-notes/1",
		Version:     "1.0.0",
	}
	if err := repo.CreateRelease(ctx, rel); err != nil {
		t.Fatalf("CreateRelease returned error: %v", err)
	}

	// Assert: Release node exists and properties mapped; relationship exists
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[rel:RELEASED]->(r:Release) RETURN r.releaseDate AS releaseDate, r.url AS url, r.version AS version",
		map[string]any{"sid": serviceID},
	)
	if err != nil {
		t.Fatalf("verify query error: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single release record: %v", err)
	}
	// Neo4j returns datetime to driver as time.Time
	gotDate, _ := rec.Get("releaseDate")
	if gd, ok := gotDate.(time.Time); !ok {
		t.Fatalf("expected time.Time for releaseDate, got %T", gotDate)
	} else if !gd.Equal(relTime) {
		t.Fatalf("expected releaseDate %v, got %v", relTime, gd)
	}
	if url, _ := rec.Get("url"); url != rel.Url {
		t.Fatalf("expected url %q, got %#v", rel.Url, url)
	}
	if ver, _ := rec.Get("version"); ver != rel.Version {
		t.Fatalf("expected version %q, got %#v", rel.Version, ver)
	}
}

func TestNeo4jReleaseRepository_CreateRelease_ServiceNotFound(t *testing.T) {
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

	rel := repositories.Release{
		ServiceId:   "00000000-0000-0000-0000-000000000000",
		ReleaseDate: time.Now().UTC(),
		Url:         "https://nowhere",
	}
	err = repo.CreateRelease(ctx, rel)
	if err == nil {
		t.Fatalf("expected error for missing service")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.Status)
	}
}
