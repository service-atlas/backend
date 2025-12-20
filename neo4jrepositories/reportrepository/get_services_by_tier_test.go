package reportrepository

import (
	"context"
	"testing"

	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/neo4jrepositories/servicerepository"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jReportRepository_GetServicesByTier_ReturnsServices(t *testing.T) {
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

	// Arrange: create services across tiers
	svc1ID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-tier-1-a",
		Description: "tier1",
		ServiceType: "api",
		Url:         "https://svc-tier-1-a",
		Tier:        1,
	})
	if err != nil {
		t.Fatalf("CreateService svc1 error: %v", err)
	}
	_ = svc1ID // not part of expected result for tier 2

	svc2ID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-tier-2-a",
		Description: "tier2-a",
		ServiceType: "worker",
		Url:         "https://svc-tier-2-a",
		Tier:        2,
	})
	if err != nil {
		t.Fatalf("CreateService svc2 error: %v", err)
	}
	svc3ID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-tier-2-b",
		Description: "tier2-b",
		ServiceType: "api",
		Url:         "https://svc-tier-2-b",
		Tier:        2,
	})
	if err != nil {
		t.Fatalf("CreateService svc3 error: %v", err)
	}

	// Act: fetch tier 2 services
	services, err := reportRepo.GetServicesByTier(ctx, 2)
	if err != nil {
		t.Fatalf("GetServicesByTier error: %v", err)
	}

	// Assert: only two services with tier 2
	if len(services) != 2 {
		t.Fatalf("expected 2 services for tier 2, got %d", len(services))
	}

	byID := map[string]repositories.Service{}
	for _, s := range services {
		byID[s.Id] = s
		if s.Tier != 2 {
			t.Errorf("expected tier 2 in results, got %d for service %+v", s.Tier, s)
		}
	}

	// Validate fields for each returned service
	if s2, ok := byID[svc2ID]; ok {
		if s2.Name != "svc-tier-2-a" || s2.Description != "tier2-a" || s2.ServiceType != "worker" || s2.Url != "https://svc-tier-2-a" || s2.Tier != 2 {
			t.Errorf("svc2 fields not mapped correctly: %+v", s2)
		}
	} else {
		t.Fatalf("service with id %s not found in results", svc2ID)
	}

	if s3, ok := byID[svc3ID]; ok {
		if s3.Name != "svc-tier-2-b" || s3.Description != "tier2-b" || s3.ServiceType != "api" || s3.Url != "https://svc-tier-2-b" || s3.Tier != 2 {
			t.Errorf("svc3 fields not mapped correctly: %+v", s3)
		}
	} else {
		t.Fatalf("service with id %s not found in results", svc3ID)
	}
}

func TestNeo4jReportRepository_GetServicesByTier_EmptyWhenNoMatch(t *testing.T) {
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

	// Arrange: create only tier 1 services
	_, err = svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-only-tier-1",
		Description: "tier1",
		ServiceType: "api",
		Url:         "https://svc-only-tier-1",
		Tier:        1,
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Act: query for a tier with no services (e.g., 3)
	services, err := reportRepo.GetServicesByTier(ctx, 3)
	if err != nil {
		t.Fatalf("GetServicesByTier error: %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected 0 services for unmatched tier, got %d", len(services))
	}
}
