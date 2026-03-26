package servicerepository

import (
	"context"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"
	"strings"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func TestNeo4jServiceRepository_CreateService_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := context.Background()

	// Start Neo4j test container
	tc, err := nRepo.NewTestContainerHelper(ctx)
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

	// Act: create service
	input := repositories.Service{
		Name:             "svc-create",
		Description:      "created service",
		ServiceType:      "api",
		Url:              "https://svc-create",
		Tier:             1,
		ArchitectureRole: "application",
		Exposure:         "public",
		ImpactDomain:     []string{"revenue", "security"},
	}
	id, err := repo.CreateService(ctx, input)
	if err != nil {
		t.Fatalf("CreateService returned error: %v", err)
	}
	if id == "" {
		t.Fatalf("expected non-empty id from CreateService")
	}

	// Assert: node exists with expected properties
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (s:Service {id: $id}) RETURN s.name AS name, s.type AS type, s.description AS description, s.url AS url, s.tier AS tier, s.created AS created, s.architecture_role AS architecture_role, s.exposure AS exposure, s.impact_domain AS impact_domain",
		map[string]any{"id": id},
	)
	if err != nil {
		t.Fatalf("failed to verify created service: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying service: %v", err)
	}
	if name, _ := rec.Get("name"); name != input.Name {
		t.Fatalf("expected name %q, got %q", input.Name, name)
	}
	if typ, _ := rec.Get("type"); typ != strings.ToUpper(input.ServiceType) {
		t.Fatalf("expected type %q, got %q", strings.ToUpper(input.ServiceType), typ)
	}
	if desc, _ := rec.Get("description"); desc != input.Description {
		t.Fatalf("expected description %q, got %q", input.Description, desc)
	}
	if url, _ := rec.Get("url"); url != input.Url {
		t.Fatalf("expected url %q, got %q", input.Url, url)
	}
	if created, _ := rec.Get("created"); created == nil {
		t.Fatalf("expected non-nil created, got %#v", created)
	}
	if crit, ok := rec.Get("tier"); ok {
		if int(crit.(int64)) != input.Tier {
			t.Fatalf("expected tier %d, got %d", input.Tier, crit)
		}
	}
	if role, _ := rec.Get("architecture_role"); role != input.ArchitectureRole {
		t.Fatalf("expected architecture_role %q, got %v", input.ArchitectureRole, role)
	}
	if exposure, _ := rec.Get("exposure"); exposure != input.Exposure {
		t.Fatalf("expected exposure %q, got %v", input.Exposure, exposure)
	}
	if domain, _ := rec.Get("impact_domain"); domain != nil {
		domainList := domain.([]any)
		if len(domainList) != len(input.ImpactDomain) {
			t.Fatalf("expected impact_domain length %d, got %d", len(input.ImpactDomain), len(domainList))
		}
		for i, d := range domainList {
			if d.(string) != input.ImpactDomain[i] {
				t.Fatalf("expected impact_domain[%d] %q, got %q", i, input.ImpactDomain[i], d)
			}
		}
	} else {
		t.Fatalf("expected non-nil impact_domain")
	}
}
