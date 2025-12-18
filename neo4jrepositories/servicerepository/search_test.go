package servicerepository

import (
	"context"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jServiceRepository_Search(t *testing.T) {
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

	err = nRepo.Startup(ctx, driver)
	if err != nil {
		t.Fatal(err)
	}
	repo := New(driver)
	id, err := repo.CreateService(ctx, repositories.Service{
		Name:        "find this test",
		Description: "a description",
		ServiceType: "service",
		Url:         "https://svc-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	services, err := repo.Search(ctx, "find")
	if err != nil {
		t.Fatal(err)
	}
	if len(services) != 1 {
		t.Fatal("expected 1 service, got", len(services))
	}
	if services[0].Id != id {
		t.Error("expected service with id", id, "got", services[0].Id)
	}
}
