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

func TestNeo4jReleaseRepository_GetReleasesByServiceId_SuccessAndPagination(t *testing.T) {
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

	// Arrange: create service and three releases with different dates
	serviceID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx, "CREATE (s:Service {id: $id, name: 'svc', type: 'api'}) RETURN s", map[string]any{"id": serviceID}); err != nil {
		t.Fatalf("create service: %v", err)
	}

	releases := []repositories.Release{
		{ServiceId: serviceID, ReleaseDate: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC), Url: "https://notes/2", Version: "2.0.0"},
		{ServiceId: serviceID, ReleaseDate: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), Url: "", Version: "1.1.0"},              // no url
		{ServiceId: serviceID, ReleaseDate: time.Date(2023, 12, 31, 23, 59, 0, 0, time.UTC), Url: "https://notes/1", Version: ""}, // no version
	}
	for _, r := range releases {
		if err := repo.CreateRelease(ctx, r); err != nil {
			t.Fatalf("CreateRelease arrange error: %v", err)
		}
	}

	// Act: page 1 size 2 should return newest two releases (Jan 2, Jan 1)
	page1, err := repo.GetReleasesByServiceId(ctx, serviceID, 1, 2)
	if err != nil {
		t.Fatalf("GetReleasesByServiceId page1 error: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 releases on page1, got %d", len(page1))
	}
	if !page1[0].ReleaseDate.Equal(releases[0].ReleaseDate) {
		t.Fatalf("expected first to be %v, got %v", releases[0].ReleaseDate, page1[0].ReleaseDate)
	}
	if page1[0].Url != "https://notes/2" || page1[0].Version != "2.0.0" {
		t.Fatalf("unexpected fields for first: %+v", page1[0])
	}
	if !page1[1].ReleaseDate.Equal(releases[1].ReleaseDate) {
		t.Fatalf("expected second to be %v, got %v", releases[1].ReleaseDate, page1[1].ReleaseDate)
	}
	// optional fields mapping
	if page1[1].Url != "" || page1[1].Version != "1.1.0" {
		t.Fatalf("unexpected fields mapping on second: %+v", page1[1])
	}

	// Act: page 2 size 2 should return the last release (Dec 31)
	page2, err := repo.GetReleasesByServiceId(ctx, serviceID, 2, 2)
	if err != nil {
		t.Fatalf("GetReleasesByServiceId page2 error: %v", err)
	}
	if len(page2) != 1 {
		t.Fatalf("expected 1 release on page2, got %d", len(page2))
	}
	if !page2[0].ReleaseDate.Equal(releases[2].ReleaseDate) {
		t.Fatalf("expected page2 item to be %v, got %v", releases[2].ReleaseDate, page2[0].ReleaseDate)
	}
	if page2[0].Url != "https://notes/1" || page2[0].Version != "" {
		t.Fatalf("unexpected fields on page2 item: %+v", page2[0])
	}
}

func TestNeo4jReleaseRepository_GetReleasesByServiceId_NotFound(t *testing.T) {
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

	_, err = repo.GetReleasesByServiceId(ctx, "00000000-0000-0000-0000-000000000000", 1, 10)
	if err == nil {
		t.Fatalf("expected error for non-existent service")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.Status)
	}
}

func TestNeo4jReleaseRepository_GetReleasesByServiceId_BadParams(t *testing.T) {
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

	// page <= 0
	_, err = repo.GetReleasesByServiceId(ctx, "some-id", 0, 10)
	if err == nil {
		t.Fatalf("expected error for page <= 0")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 400 {
		t.Fatalf("expected HTTP 400, got %d", httpErr.Status)
	}

	// pageSize <= 0
	_, err = repo.GetReleasesByServiceId(ctx, "some-id", 1, 0)
	if err == nil {
		t.Fatalf("expected error for pageSize <= 0")
	}
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 400 {
		t.Fatalf("expected HTTP 400, got %d", httpErr.Status)
	}
}
