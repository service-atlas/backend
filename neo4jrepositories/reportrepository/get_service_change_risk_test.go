package reportrepository

import (
	"context"
	"errors"
	"testing"

	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/neo4jrepositories/servicerepository"
	"service-atlas/repositories"

	"service-atlas/internal/customerrors"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jReportRepository_GetServiceChangeRisk_NotFound(t *testing.T) {
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
	_, err = repo.GetServiceChangeRisk(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatalf("expected error for non-existent service")
	}
	if httpErr, ok := errors.AsType[*customerrors.HTTPError](err); !ok || httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404 error, got %T: %v", err, err)
	}
}

func TestNeo4jReportRepository_GetServiceChangeRisk_LowMediumHigh(t *testing.T) {
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

	// LOW: tier 4 with no dependents => base 5
	lowID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "low-svc",
		Description: "low",
		ServiceType: "api",
		Url:         "https://low",
		Tier:        4,
	})
	if err != nil {
		t.Fatalf("CreateService low error: %v", err)
	}

	// MEDIUM: tier 3 base 10 + dependents (t2 + t2 + t3) => 10 + 12 + 12 + 5 = 39
	medTarget, err := svcRepo.CreateService(ctx, repositories.Service{Name: "med-target", ServiceType: "api", Url: "https://med-target", Tier: 3})
	if err != nil {
		t.Fatalf("CreateService med target error: %v", err)
	}
	medDep2a, _ := svcRepo.CreateService(ctx, repositories.Service{Name: "med-dep2a", ServiceType: "worker", Url: "https://dep2a", Tier: 2})
	medDep2b, _ := svcRepo.CreateService(ctx, repositories.Service{Name: "med-dep2b", ServiceType: "worker", Url: "https://dep2b", Tier: 2})
	medDep3, _ := svcRepo.CreateService(ctx, repositories.Service{Name: "med-dep3", ServiceType: "worker", Url: "https://dep3", Tier: 3})

	// HIGH: tier 1 base 40 + (t2 + t1) => 40 + 12 + 20 = 72
	highTarget, err := svcRepo.CreateService(ctx, repositories.Service{Name: "high-target", ServiceType: "api", Url: "https://high-target", Tier: 1})
	if err != nil {
		t.Fatalf("CreateService high target error: %v", err)
	}
	highDep2, _ := svcRepo.CreateService(ctx, repositories.Service{Name: "high-dep2", ServiceType: "worker", Url: "https://high-dep2", Tier: 2})
	highDep1, _ := svcRepo.CreateService(ctx, repositories.Service{Name: "high-dep1", ServiceType: "worker", Url: "https://high-dep1", Tier: 1})

	write := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() { _ = write.Close(ctx) }()

	// create dependencies for MEDIUM target
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": medDep2a, "b": medTarget}); err != nil {
		t.Fatalf("create med dep2a -> target: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": medDep2b, "b": medTarget}); err != nil {
		t.Fatalf("create med dep2b -> target: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": medDep3, "b": medTarget}); err != nil {
		t.Fatalf("create med dep3 -> target: %v", err)
	}

	// create dependencies for HIGH target
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": highDep2, "b": highTarget}); err != nil {
		t.Fatalf("create high dep2 -> target: %v", err)
	}
	if _, err = write.Run(ctx, "MATCH (a:Service {id: $a}),(b:Service {id: $b}) MERGE (a)-[:DEPENDS_ON]->(b)", map[string]any{"a": highDep1, "b": highTarget}); err != nil {
		t.Fatalf("create high dep1 -> target: %v", err)
	}

	// Act & Assert
	lowRisk, err := reportRepo.GetServiceChangeRisk(ctx, lowID)
	if err != nil {
		t.Fatalf("GetServiceChangeRisk low error: %v", err)
	}
	if lowRisk.Risk != "low" {
		t.Fatalf("expected low risk, got %s (score=%d)", lowRisk.Risk, lowRisk.Score)
	}

	medRisk, err := reportRepo.GetServiceChangeRisk(ctx, medTarget)
	if err != nil {
		t.Fatalf("GetServiceChangeRisk medium error: %v", err)
	}
	if medRisk.Risk != "medium" {
		t.Fatalf("expected medium risk, got %s (score=%d)", medRisk.Risk, medRisk.Score)
	}

	highRisk, err := reportRepo.GetServiceChangeRisk(ctx, highTarget)
	if err != nil {
		t.Fatalf("GetServiceChangeRisk high error: %v", err)
	}
	if highRisk.Risk != "high" {
		t.Fatalf("expected high risk, got %s (score=%d)", highRisk.Risk, highRisk.Score)
	}
}
