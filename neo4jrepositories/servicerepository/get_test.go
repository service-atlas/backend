package servicerepository

import (
	"context"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jServiceRepository_GetServiceById_Success(t *testing.T) {
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

	// Arrange: create a service using repository
	createdID, err := repo.CreateService(ctx, repositories.Service{
		Name:        "svc-get",
		Description: "desc",
		ServiceType: "worker",
		Url:         "https://svc-get",
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Act
	svc, err := repo.GetServiceById(ctx, createdID)
	if err != nil {
		t.Fatalf("GetServiceById error: %v", err)
	}

	// Assert
	if svc.Id != createdID {
		t.Fatalf("expected id %s, got %s", createdID, svc.Id)
	}
	if svc.Name != "svc-get" || svc.Description != "desc" || svc.ServiceType != "worker" || svc.Url != "https://svc-get" {
		t.Fatalf("fields not mapped correctly: %+v", svc)
	}
}

func TestNeo4jServiceRepository_GetServiceById_NotFound(t *testing.T) {
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

	svc, err := repo.GetServiceById(ctx, "00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("GetServiceById returned error: %v", err)
	}
	// Code returns zero-value when not found
	if svc.Id != "" || svc.Name != "" {
		t.Fatalf("expected zero-value service when not found, got: %+v", svc)
	}
}
