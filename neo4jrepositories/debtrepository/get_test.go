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

func TestNeo4jDebtRepository_GetDebtByServiceId_BasicAndFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()

	// Start Neo4j test container
	tc, err := neo4jrepositories.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(context.Background()) })

	// Connect driver
	driver, err := neo4j.NewDriverWithContext(tc.Endpoint, neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	repo := New(driver)

	// Arrange: create Service
	serviceID := "55555555-5555-5555-5555-555555555555"
	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()
	if _, err = write.Run(ctx,
		"CREATE (s:Service {id: $sid, name: $sname}) RETURN s",
		map[string]any{"sid": serviceID, "sname": "get-test-service"},
	); err != nil {
		t.Fatalf("failed to create service node: %v", err)
	}

	// Create two debts via repository (status will be DefaultStatus)
	d1 := repositories.Debt{Type: "debt", Title: "Debt One", Description: "first", ServiceId: serviceID}
	d2 := repositories.Debt{Type: "debt", Title: "Debt Two", Description: "second", ServiceId: serviceID}
	if err := repo.CreateDebtItem(ctx, d1); err != nil {
		t.Fatalf("CreateDebtItem(d1) error: %v", err)
	}
	if err := repo.CreateDebtItem(ctx, d2); err != nil {
		t.Fatalf("CreateDebtItem(d2) error: %v", err)
	}

	// Fetch IDs back for further checks
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[:OWNS]->(d:Debt) RETURN collect({id:d.id,title:d.title,status:d.status}) AS debts",
		map[string]any{"sid": serviceID},
	)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record: %v", err)
	}
	vals, ok := rec.Get("debts")
	if !ok {
		t.Fatalf("missing 'debts' property")
	}
	arr, ok := vals.([]any)
	if !ok {
		t.Fatalf("expected array, got %T", vals)
	}
	if len(arr) < 2 {
		t.Fatalf("expected at least 2 debts, got %d", len(arr))
	}
	// Update one to remediated using repository UpdateStatus through a fresh repo to mirror prod usage
	updRepo := New(driver)
	// pull one id
	first, ok := arr[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", arr[0])
	}
	firstID, ok := first["id"].(string)
	if !ok {
		t.Fatalf("expected string, got %T", first["id"])
	}
	if err := updRepo.UpdateStatus(ctx, firstID, "remediated"); err != nil {
		t.Fatalf("UpdateStatus error: %v", err)
	}

	// Act: Get without filter (should return both)
	list, err := repo.GetDebtByServiceId(ctx, serviceID, 1, 10, false)
	if err != nil {
		t.Fatalf("GetDebtByServiceId error: %v", err)
	}
	if len(list) < 2 {
		t.Fatalf("expected at least 2 debts, got %d", len(list))
	}
	for _, d := range list {
		if d.ServiceId != serviceID {
			t.Errorf("expected ServiceId %s, got %s", serviceID, d.ServiceId)
		}
		if d.Id == "" {
			t.Errorf("expected non-empty id")
		}
		if d.Title == "" {
			t.Errorf("expected non-empty title")
		}
		if d.Status == "" {
			t.Errorf("expected non-empty status")
		}
	}

	// Act: Get only resolved
	resolved, err := repo.GetDebtByServiceId(ctx, serviceID, 1, 10, true)
	if err != nil {
		t.Fatalf("GetDebtByServiceId(onlyResolved) error: %v", err)
	}
	if len(resolved) != 1 {
		t.Fatalf("expected exactly 1 resolved debt, got %d", len(resolved))
	}
	if resolved[0].Status != "remediated" {
		t.Errorf("expected status 'remediated', got %q", resolved[0].Status)
	}
}

func TestNeo4jDebtRepository_GetDebtByServiceId_InvalidPaging(t *testing.T) {
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

	_, err = repo.GetDebtByServiceId(ctx, "some-service", 0, 10, false)
	if err == nil {
		t.Fatalf("expected error for invalid page")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 400 {
		t.Fatalf("expected 400, got %d", httpErr.Status)
	}

	_, err = repo.GetDebtByServiceId(ctx, "some-service", 1, 0, false)
	if err == nil {
		t.Fatalf("expected error for invalid page size")
	}
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 400 {
		t.Fatalf("expected 400, got %d", httpErr.Status)
	}
}
