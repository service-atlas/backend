package reportrepository

import (
	"context"
	"testing"

	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/neo4jrepositories/servicerepository"
	"service-atlas/neo4jrepositories/teamrepository"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jReportRepository_GetServicesByTeam_ReturnsServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()
	// spin up test container
	tc, err := nRepo.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(ctx) })

	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	teamRepo := teamrepository.New(driver)
	svcRepo := servicerepository.New(driver)
	reportRepo := New(driver)

	// Arrange: create a team
	teamID, err := teamRepo.CreateTeam(ctx, repositories.Team{Name: "team-a"})
	if err != nil {
		t.Fatalf("CreateTeam error: %v", err)
	}
	// Arrange: create two services
	svc1ID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-1",
		Description: "first",
		ServiceType: "api",
		Url:         "https://svc-1",
	})
	if err != nil {
		t.Fatalf("CreateService 1 error: %v", err)
	}
	svc2ID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-2",
		Description: "second",
		ServiceType: "worker",
		Url:         "https://svc-2",
	})
	if err != nil {
		t.Fatalf("CreateService 2 error: %v", err)
	}
	// Arrange: associate team -> services
	if err := teamRepo.CreateTeamAssociation(ctx, teamID, svc1ID); err != nil {
		t.Fatalf("CreateTeamAssociation 1 error: %v", err)
	}
	if err := teamRepo.CreateTeamAssociation(ctx, teamID, svc2ID); err != nil {
		t.Fatalf("CreateTeamAssociation 2 error: %v", err)
	}

	// Act
	services, err := reportRepo.GetServicesByTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("GetServicesByTeam error: %v", err)
	}

	// Assert
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}
	// make map by id for stable assertions
	byID := map[string]repositories.Service{}
	for _, s := range services {
		byID[s.Id] = s
	}
	// svc1
	s1, ok := byID[svc1ID]
	if !ok {
		t.Fatalf("service with id %s not found in result", svc1ID)
	}
	if s1.Name != "svc-1" || s1.Description != "first" || s1.ServiceType != "Api" || s1.Url != "https://svc-1" {
		t.Errorf("svc1 fields not mapped correctly: %+v", s1)
	}
	// svc2
	s2, ok := byID[svc2ID]
	if !ok {
		t.Fatalf("service with id %s not found in result", svc2ID)
	}
	if s2.Name != "svc-2" || s2.Description != "second" || s2.ServiceType != "Worker" || s2.Url != "https://svc-2" {
		t.Errorf("svc2 fields not mapped correctly: %+v", s2)
	}
}

func TestNeo4jReportRepository_GetServicesByTeam_EmptyWhenNoServices(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()
	tc, err := nRepo.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(ctx) })

	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	reportRepo := New(driver)
	teamRepo := teamrepository.New(driver)

	teamID, err := teamRepo.CreateTeam(ctx, repositories.Team{Name: "lonely-team"})
	if err != nil {
		t.Fatalf("CreateTeam error: %v", err)
	}

	services, err := reportRepo.GetServicesByTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("GetServicesByTeam error: %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected 0 services, got %d", len(services))
	}
}

func TestNeo4jReportRepository_GetServicesByTeam_EmptyWhenTeamNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()
	tc, err := nRepo.NewTestContainerHelper(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tc.Container.Terminate(ctx) })

	driver, err := neo4j.NewDriverWithContext(
		tc.Endpoint,
		neo4j.BasicAuth("neo4j", "letmein!", ""))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = driver.Close(ctx) }()

	reportRepo := New(driver)
	// Act: use some random id that doesn't exist
	services, err := reportRepo.GetServicesByTeam(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetServicesByTeam error: %v", err)
	}
	if len(services) != 0 {
		t.Fatalf("expected 0 services for non-existent team, got %d", len(services))
	}
}
