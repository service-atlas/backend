package debtrepository

import (
	"context"
	"errors"
	"service-atlas/internal/customerrors"
	"service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jDebtRepository_CreateDebtItem(t *testing.T) {
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

	// Arrange: create Service
	serviceID := "44444444-4444-4444-4444-444444444444"

	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()

	if _, err = write.Run(ctx,
		"CREATE (s:Service {id: $sid, name: $sname}) RETURN s",
		map[string]any{"sid": serviceID, "sname": "test-service"},
	); err != nil {
		t.Fatalf("failed to create service node: %v", err)
	}

	d := repositories.Debt{
		Type:        "debt",
		Title:       "Test Debt",
		Description: "This is a test debt",
		Status:      "pending", // ignored by create.go; DefaultStatus is used instead
		ServiceId:   serviceID,
	}

	// Act
	if err := repo.CreateDebtItem(ctx, d); err != nil {
		t.Fatalf("CreateDebtItem returned error: %v", err)
	}

	// Assert: debt node and relationship exist
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()

	res, err := read.Run(ctx,
		"MATCH (s:Service {id: $sid})-[:OWNS]->(d:Debt) RETURN d.id as id, d.created as created, d.title as title, d.description as description, d.status as status, d.type as type",
		map[string]any{"sid": serviceID},
	)
	if err != nil {
		t.Fatalf("failed to verify debt creation: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil || rec == nil {
		t.Fatalf("expected single record verifying debt creation, got err=%v", err)
	}
	if ty, _ := rec.Get("type"); ty != d.Type {
		t.Errorf("expected type %q, got %q", d.Type, ty)
	}
	if title, _ := rec.Get("title"); title != d.Title {
		t.Errorf("expected title %q, got %q", d.Title, title)
	}
	if desc, _ := rec.Get("description"); desc != d.Description {
		t.Errorf("expected description %q, got %q", d.Description, desc)
	}
	if status, _ := rec.Get("status"); status != DefaultStatus {
		t.Errorf("expected status %q, got %q", DefaultStatus, status)
	}
	if id, _ := rec.Get("id"); id == nil || id == "" {
		t.Errorf("expected auto-generated id, got %#v", id)
	}
	if created, _ := rec.Get("created"); created == nil {
		t.Errorf("expected non-nil created, got %#v", created)
	}
}

func TestNeo4jDebtRepository_CreateDebtItem_NoService(t *testing.T) {
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

	d := repositories.Debt{
		Type:        "debt",
		Title:       "Missing Service Debt",
		Description: "Should fail",
		ServiceId:   "00000000-0000-0000-0000-000000000000", // not created on purpose
	}

	err = repo.CreateDebtItem(ctx, d)
	if err == nil {
		t.Fatalf("expected error when service does not exist")
	}

	// Check it is our HTTPError with 404
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d (msg=%q)", httpErr.Status, httpErr.Msg)
	}
}
