package servicerepository

import (
	"context"
	"errors"
	"service-atlas/internal/auth"
	"service-atlas/internal/customerrors"
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

func TestNeo4jServiceRepository_CreateService_WithCreatedBy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := auth.ContextWithName(context.Background(), "John Doe")

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
		Name:             "svc-created-by",
		Description:      "created service with creator",
		ServiceType:      "api",
		Tier:             1,
		ArchitectureRole: "application",
		Exposure:         "public",
	}
	id, err := repo.CreateService(ctx, input)
	if err != nil {
		t.Fatalf("CreateService returned error: %v", err)
	}

	// Assert: node exists with expected properties
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (s:Service {id: $id}) RETURN s.createdBy AS createdBy",
		map[string]any{"id": id},
	)
	if err != nil {
		t.Fatalf("failed to verify created service: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying service: %v", err)
	}
	if cb, _ := rec.Get("createdBy"); cb != "John Doe" {
		t.Fatalf("expected createdBy %q, got %q", "John Doe", cb)
	}
}

func TestUpdateService(t *testing.T) {
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

	// Arrange: create a service
	createdID, err := repo.CreateService(ctx, repositories.Service{
		Name:        "svc-update",
		Description: "before",
		ServiceType: "api",
		Url:         "https://before",
		Tier:        1,
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Act: update service fields
	u := repositories.Service{
		Id:               createdID,
		Name:             "svc-updated",
		Description:      "after",
		ServiceType:      "worker",
		Url:              "https://after",
		Tier:             2,
		ArchitectureRole: "infrastructure",
		Exposure:         "private",
		ImpactDomain:     []string{"compliance"},
	}
	if err := repo.UpdateService(ctx, u); err != nil {
		t.Fatalf("UpdateService returned error: %v", err)
	}

	// Assert: fields updated and updated timestamp set
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (s:Service {id: $id}) RETURN s.name AS name, s.type AS type, s.description AS description, s.url AS url, s.tier AS tier, s.updated AS updated, s.architecture_role AS architecture_role, s.exposure AS exposure, s.impact_domain AS impact_domain",
		map[string]any{"id": createdID},
	)
	if err != nil {
		t.Fatalf("failed to verify updated service: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying update: %v", err)
	}
	if name, _ := rec.Get("name"); name != u.Name {
		t.Fatalf("expected name %q, got %q", u.Name, name)
	}
	if typ, _ := rec.Get("type"); typ != strings.ToUpper(u.ServiceType) {
		t.Fatalf("expected type %q, got %q", strings.ToUpper(u.ServiceType), typ)
	}
	if desc, _ := rec.Get("description"); desc != u.Description {
		t.Fatalf("expected description %q, got %q", u.Description, desc)
	}
	if url, _ := rec.Get("url"); url != u.Url {
		t.Fatalf("expected url %q, got %q", u.Url, url)
	}
	if upd, _ := rec.Get("updated"); upd == nil {
		t.Fatalf("expected non-nil updated timestamp, got %#v", upd)
	}
	if crit, ok := rec.Get("tier"); ok {
		critInt64, ok := crit.(int64)
		if !ok {
			t.Fatalf("expected int64 tier from neo4j, got %T: %v", crit, crit)
		}
		if critInt := int(critInt64); critInt != u.Tier {
			t.Fatalf("expected tier %d, got %d", u.Tier, critInt)
		}
	}
	if role, _ := rec.Get("architecture_role"); role != u.ArchitectureRole {
		t.Fatalf("expected architecture_role %q, got %v", u.ArchitectureRole, role)
	}
	if exposure, _ := rec.Get("exposure"); exposure != u.Exposure {
		t.Fatalf("expected exposure %q, got %v", u.Exposure, exposure)
	}
	if domain, _ := rec.Get("impact_domain"); domain != nil {
		domainList := domain.([]any)
		if len(domainList) != len(u.ImpactDomain) {
			t.Fatalf("expected impact_domain length %d, got %d", len(u.ImpactDomain), len(domainList))
		}
		for i, d := range domainList {
			if d.(string) != u.ImpactDomain[i] {
				t.Fatalf("expected impact_domain[%d] %q, got %q", i, u.ImpactDomain[i], d)
			}
		}
	} else {
		t.Fatalf("expected non-nil impact_domain")
	}
}

func TestUpdateService_WithUpdatedBy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	ctx := auth.ContextWithName(context.Background(), "Jane Doe")

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

	// Arrange: create a service
	createdID, err := repo.CreateService(ctx, repositories.Service{
		Name:        "svc-update-by",
		Description: "before",
		ServiceType: "api",
		Tier:        1,
	})
	if err != nil {
		t.Fatalf("CreateService error: %v", err)
	}

	// Act: update service fields
	u := repositories.Service{
		Id:               createdID,
		Name:             "svc-updated-by",
		Description:      "after",
		ServiceType:      "worker",
		Tier:             2,
		ArchitectureRole: "infrastructure",
	}
	if err := repo.UpdateService(ctx, u); err != nil {
		t.Fatalf("UpdateService returned error: %v", err)
	}

	// Assert: updatedBy set
	read := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = read.Close(ctx) }()
	res, err := read.Run(ctx,
		"MATCH (s:Service {id: $id}) RETURN s.updatedBy AS updatedBy",
		map[string]any{"id": createdID},
	)
	if err != nil {
		t.Fatalf("failed to verify updated service: %v", err)
	}
	rec, err := res.Single(ctx)
	if err != nil {
		t.Fatalf("expected single record verifying update: %v", err)
	}
	if ub, _ := rec.Get("updatedBy"); ub != "Jane Doe" {
		t.Fatalf("expected updatedBy %q, got %q", "Jane Doe", ub)
	}
}

func TestNeo4jServiceRepository_UpdateService_NotFound(t *testing.T) {
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

	err = repo.UpdateService(ctx, repositories.Service{Id: "00000000-0000-0000-0000-000000000000", Name: "x", ServiceType: "api", Url: "https://x"})
	if err == nil {
		t.Fatalf("expected error when service not found")
	}
	var httpErr *customerrors.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *customerrors.HTTPError, got %T: %v", err, err)
	}
	if httpErr.Status != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.Status)
	}
}
