package servicerepository

import (
	"context"
	nRepo "service-atlas/neo4jrepositories"
	teamrepo "service-atlas/neo4jrepositories/teamrepository"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jServiceRepository_GetTeamsByServiceId_Success(t *testing.T) {
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

	svcRepo := New(driver)
	tRepo := teamrepo.New(driver)

	// Arrange: create a service
	serviceID, err := svcRepo.CreateService(ctx, repositories.Service{
		Name:        "svc-has-team",
		Description: "desc",
		ServiceType: "api",
		Url:         "https://svc-has-team",
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Arrange: create a team
	teamID, err := tRepo.CreateTeam(ctx, repositories.Team{Name: "Team A"})
	if err != nil {
		t.Fatalf("CreateTeam error: %v", err)
	}

	// Arrange: associate team -> service
	if err := tRepo.CreateTeamAssociation(ctx, teamID, serviceID); err != nil {
		t.Fatalf("CreateTeamAssociation error: %v", err)
	}

	// Act
	teams, err := svcRepo.GetTeamsByServiceId(ctx, serviceID)
	if err != nil {
		t.Fatalf("GetTeamsByServiceId error: %v", err)
	}

	// Assert
	if len(teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(teams))
	}
	if teams[0].Id == "" || teams[0].Name != "Team A" {
		t.Fatalf("unexpected team returned: %+v", teams[0])
	}
}

func TestNeo4jServiceRepository_GetTeamsByServiceId_NotFound(t *testing.T) {
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

	svcRepo := New(driver)

	// Act: use a random/non-existent id
	teams, err := svcRepo.GetTeamsByServiceId(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(teams) != 0 {
		t.Fatalf("expected 0 teams, got %d", len(teams))
	}
}
