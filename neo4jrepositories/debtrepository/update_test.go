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

func TestNeo4jDebtRepository_UpdateStatus_Success(t *testing.T) {
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

	// Arrange: create Service and a Debt
	serviceID := "66666666-6666-6666-6666-666666666666"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx,
		"CREATE (s:Service {id: $sid, name: $sname}) RETURN s",
		map[string]any{"sid": serviceID, "sname": "update-test-service"},
	); err != nil {
		t.Fatalf("failed to create service node: %v", err)
	}

	deb := repositories.Debt{Type: "debt", Title: "Needs Update", Description: "status change", ServiceId: serviceID}
	if err := repo.CreateDebtItem(ctx, deb); err != nil {
		t.Fatalf("CreateDebtItem error: %v", err)
	}

	// Fetch the created Debt id
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[:OWNS]->(d:Debt) RETURN d.id AS id, d.status AS status LIMIT 1",
		map[string]any{"sid": serviceID},
	)
	if err != nil {
		t.Fatalf("failed to fetch created debt: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record: %v", err)
	}
	idVal, _ := rec.Get("id")
	debtID := idVal.(string)

	// Act: update status
	newStatus := "remediated"
	if err := repo.UpdateStatus(ctx, debtID, newStatus); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}

	// Assert: status updated in DB
	check, err := read.Run(ctx,
		"MATCH (d:Debt {id: $id}) RETURN d.status AS status",
		map[string]any{"id": debtID},
	)
	if err != nil {
		t.Fatalf("failed to verify updated status: %v", err)
	}
	rec2, err := check.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying status: %v", err)
	}
	st, _ := rec2.Get("status")
	if st != newStatus {
		t.Fatalf("expected status %q, got %q", newStatus, st)
	}
}

func TestNeo4jDebtRepository_UpdateStatus_NotFound(t *testing.T) {
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

	err = repo.UpdateStatus(ctx, "00000000-0000-0000-0000-000000000000", "remediated")
	if err == nil {
		t.Fatalf("expected error when debt does not exist")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.Status)
	}
}
