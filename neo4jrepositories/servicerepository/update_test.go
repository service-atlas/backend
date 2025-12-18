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

func TestNeo4jServiceRepository_UpdateService_Success(t *testing.T) {
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
		Name:        "svc-update",
		Description: "before",
		ServiceType: "api",
		Url:         "https://before",
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Act: update service fields
	u := repositories.Service{
		Id:          createdID,
		Name:        "svc-updated",
		Description: "after",
		ServiceType: "worker",
		Url:         "https://after",
	}
	if err := repo.UpdateService(ctx, u); err != nil {
		t.Fatalf("UpdateService returned error: %v", err)
	}

	// Assert: fields updated and updated timestamp set
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (s:Service {id: $id}) RETURN s.name AS name, s.type AS type, s.description AS description, s.url AS url, s.updated AS updated",
		map[string]any{"id": createdID},
	)
	if err != nil {
		t.Fatalf("failed to verify updated service: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying update: %v", err)
	}
	if name, _ := rec.Get("name"); name != u.Name {
		t.Fatalf("expected name %q, got %q", u.Name, name)
	}
	if typ, _ := rec.Get("type"); typ != u.ServiceType {
		t.Fatalf("expected type %q, got %q", u.ServiceType, typ)
	}
	if desc, _ := rec.Get("description"); desc != u.Description {
		t.Fatalf("expected description %q, got %q", u.Description, desc)
	}
	if url, _ := rec.Get("url"); url != u.Url {
		t.Fatalf("expected url %q, got %q", u.Url, url)
	}
	if upd, _ := rec.Get("updated"); upd == nil {
		t.Fatalf("expected non-nil updated timestamp, got %#v", upd)
	}
}

func TestNeo4jServiceRepository_UpdateService_NotFound(t *testing.T) {
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

	err = repo.UpdateService(ctx, repositories.Service{Id: "00000000-0000-0000-0000-000000000000", Name: "x", ServiceType: "api", Url: "https://x"})
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
