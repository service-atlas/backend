package reportrepository

import (
	"context"
	"testing"

	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/neo4jrepositories/servicerepository"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jReportRepository_GetServiceTypes(t *testing.T) {
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

	svcRepo := servicerepository.New(driver)
	reportRepo := New(driver)

	// Seed data
	services := []repositories.Service{
		{Name: "svc1", ServiceType: "api", Url: "https://svc1", Tier: 1},
		{Name: "svc2", ServiceType: "api", Url: "https://svc2", Tier: 2},
		{Name: "svc3", ServiceType: "worker", Url: "https://svc3", Tier: 3},
		{Name: "svc4", ServiceType: "database", Url: "https://svc4", Tier: 4},
		{Name: "svc5", ServiceType: "worker", Url: "https://svc5", Tier: 3},
		{Name: "svc6", ServiceType: "worker", Url: "https://svc6", Tier: 3},
	}

	for _, s := range services {
		_, err := svcRepo.CreateService(ctx, s)
		if err != nil {
			t.Fatalf("failed to create service %s: %v", s.Name, err)
		}
	}

	// Act
	actual, err := reportRepo.GetServiceTypes(ctx)
	if err != nil {
		t.Fatalf("GetServiceTypes error: %v", err)
	}

	// Expected results:
	// worker: 3
	// api: 2
	// database: 1
	// Note: GetServiceTypes uses internal.ToTitleCase, so "worker" -> "Worker", "api" -> "Api", "database" -> "Database"

	if len(actual) != 3 {
		t.Errorf("expected 3 service types, got %d", len(actual))
	}

	expected := map[string]int64{
		"Worker":   3,
		"Api":      2,
		"Database": 1,
	}

	for _, rt := range actual {
		count, ok := expected[rt.Type]
		if !ok {
			t.Errorf("unexpected service type: %s", rt.Type)
			continue
		}
		if rt.Count != count {
			t.Errorf("expected count %d for type %s, got %d", count, rt.Type, rt.Count)
		}
		delete(expected, rt.Type)
	}

	if len(expected) > 0 {
		t.Errorf("missing service types: %v", expected)
	}
}
