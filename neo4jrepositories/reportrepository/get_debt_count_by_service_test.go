package reportrepository

import (
	"context"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/neo4jrepositories/debtrepository"
	"service-atlas/neo4jrepositories/servicerepository"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Test that GetDebtCountByService:
// - counts only debts with status in ("in_progress", "pending")
// - excludes services with zero qualifying debts
func TestNeo4jReportRepository_GetDebtCountByService_FilterAndExclude(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()

	// Spin up Neo4j test container
	tc, err := nRepo.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(context.Background()) })

	// Connect driver
	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	reportRepo := New(driver)
	svcRepo := servicerepository.New(driver)
	debtRepo := debtrepository.New(driver)

	// Arrange: create services
	svcA, err := svcRepo.CreateService(ctx, repositories.Service{Name: "svc-A", ServiceType: "api", Url: "https://a"})
	if err != nil {
		t.Fatalf("create svcA: %v", err)
	}
	svcB, err := svcRepo.CreateService(ctx, repositories.Service{Name: "svc-B", ServiceType: "worker", Url: "https://b"})
	if err != nil {
		t.Fatalf("create svcB: %v", err)
	}
	_, err = svcRepo.CreateService(ctx, repositories.Service{Name: "svc-C", ServiceType: "db", Url: "https://c"}) // no debts
	if err != nil {
		t.Fatalf("create svcC: %v", err)
	}
	svcD, err := svcRepo.CreateService(ctx, repositories.Service{Name: "svc-D", ServiceType: "api", Url: "https://d"})
	if err != nil {
		t.Fatalf("create svcD: %v", err)
	}

	// Debts for A: three pending (default)
	for i := 0; i < 3; i++ {
		if err := debtRepo.CreateDebtItem(ctx, repositories.Debt{Type: "debt", Title: "A-" + string(rune('1'+i)), Description: "", ServiceId: svcA}); err != nil {
			t.Fatalf("CreateDebtItem A-%d: %v", i+1, err)
		}
	}

	// Debts for B: one pending (qualifies), and one updated to in_progress (also qualifies)
	if err := debtRepo.CreateDebtItem(ctx, repositories.Debt{Type: "debt", Title: "B-1", Description: "", ServiceId: svcB}); err != nil {
		t.Fatalf("CreateDebtItem B-1: %v", err)
	}
	if err := debtRepo.CreateDebtItem(ctx, repositories.Debt{Type: "debt", Title: "B-2", Description: "", ServiceId: svcB}); err != nil {
		t.Fatalf("CreateDebtItem B-2: %v", err)
	}

	// Fetch one of B's debts and set it to in_progress; the other remains pending
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[:OWNS]->(d:Debt) RETURN collect(d.id) AS ids",
		map[string]any{"sid": svcB},
	)
	if err != nil {
		t.Fatalf("read B debts: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("read single: %v", err)
	}
	v, _ := rec.Get("ids")
	ids, _ := v.([]any)
	if len(ids) != 2 {
		t.Fatalf("expected 2 debts for B, got %d", len(ids))
	}
	firstId, _ := ids[0].(string)
	if err := debtRepo.UpdateStatus(ctx, firstId, "in_progress"); err != nil {
		t.Fatalf("update B debt to in_progress: %v", err)
	}

	// Debts for D: one created then remediated (should be excluded)
	if err := debtRepo.CreateDebtItem(ctx, repositories.Debt{Type: "debt", Title: "D-1", Description: "", ServiceId: svcD}); err != nil {
		t.Fatalf("CreateDebtItem D-1: %v", err)
	}
	// fetch D's debt id and set to remediated
	res2, err := read.Run(ctx,
		"MATCH (:Service {id: $sid})-[:OWNS]->(d:Debt) RETURN d.id AS id",
		map[string]any{"sid": svcD},
	)
	if err != nil {
		t.Fatalf("read D debts: %v", err)
	}
	if res2.Next(ctx) {
		did, _ := res2.Record().Get("id")
		if s, ok := did.(string); ok {
			if err := debtRepo.UpdateStatus(ctx, s, "remediated"); err != nil {
				t.Fatalf("update D debt to remediated: %v", err)
			}
		} else {
			t.Fatalf("unexpected id type %T", did)
		}
	} else if err := res2.Err(); err != nil {
		t.Fatalf("iter D debts: %v", err)
	}

	// Act
	report, err := reportRepo.GetDebtCountByService(ctx)
	if err != nil {
		t.Fatalf("GetDebtCountByService error: %v", err)
	}

	// Convert to map for easy assertions
	got := make(map[string]int64)
	names := make(map[string]string)
	for _, r := range report {
		got[r.Id] = r.Count
		names[r.Id] = r.Name
	}

	// Assert: only A and B are present with correct counts (A=3, B=2)
	if len(got) != 2 {
		t.Fatalf("expected exactly 2 services in report, got %d", len(got))
	}
	if c := got[svcA]; c != 3 {
		t.Fatalf("svcA expected count=3, got %d (name=%q)", c, names[svcA])
	}
	if c := got[svcB]; c != 2 {
		t.Fatalf("svcB expected count=2, got %d (name=%q)", c, names[svcB])
	}
	if _, ok := got[svcD]; ok {
		t.Fatalf("svcD should be excluded (no qualifying debt)")
	}
}
