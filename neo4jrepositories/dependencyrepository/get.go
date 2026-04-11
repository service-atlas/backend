package dependencyrepository

import (
	"context"
	"fmt"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (d *Neo4jDependencyRepository) GetDependencies(ctx context.Context, id string) ([]*repositories.Dependency, error) {

	query := `
			MATCH (s1:Service {id: $serviceId})-[r:DEPENDS_ON]->(s2:Service)
			RETURN s2.id as id, s2.name as name, r.version as version, s2.type as type, r.interaction_type as interaction_type
		`
	parameters := map[string]any{
		"serviceId": id,
	}
	result, err := d.manager.ExecuteRead(ctx, makeGetTransaction(ctx, query, parameters))
	if err != nil {
		return nil, err
	}

	deps := result.([]*repositories.Dependency)
	if deps == nil {
		return []*repositories.Dependency{}, nil
	}
	return deps, nil
}
func (d *Neo4jDependencyRepository) GetDependenciesByInteractionType(ctx context.Context, id, interaction_type string) ([]*repositories.Dependency, error) {

	query := `
			MATCH (s1:Service {id: $serviceId})-[r:DEPENDS_ON {interaction_type: $interaction_type}]->(s2:Service)
			RETURN s2.id as id, s2.name as name, r.version as version, s2.type as type, r.interaction_type as interaction_type
		`
	parameters := map[string]any{
		"serviceId":        id,
		"interaction_type": interaction_type,
	}
	result, err := d.manager.ExecuteRead(ctx, makeGetTransaction(ctx, query, parameters))
	if err != nil {
		return nil, err
	}

	deps := result.([]*repositories.Dependency)
	if deps == nil {
		return []*repositories.Dependency{}, nil
	}
	return deps, nil
}

func (d *Neo4jDependencyRepository) GetDependents(ctx context.Context, id string) ([]*repositories.Dependency, error) {
	query := `
			MATCH (s1:Service)-[r:DEPENDS_ON]->(s2:Service {id: $serviceId})
			RETURN s1.id as id, s1.name as name, s1.type as type, r.version as version, r.interaction_type as interaction_type
		`
	parameters := map[string]any{
		"serviceId": id,
	}
	result, err := d.manager.ExecuteRead(ctx, makeGetTransaction(ctx, query, parameters))
	if err != nil {
		return nil, err
	}
	return result.([]*repositories.Dependency), nil
}

func makeGetTransaction(ctx context.Context, query string, parameters map[string]any) func(tx neo4j.ManagedTransaction) (any, error) {
	return func(tx neo4j.ManagedTransaction) (any, error) {
		// First check if the service exists
		checkQuery := `
			MATCH (s:Service {id: $serviceId})
			RETURN s
		`
		result, err := tx.Run(ctx, checkQuery, parameters)
		if err != nil {
			return nil, err
		}

		// If no records are returned, the service doesn't exist
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			serviceId, ok := parameters["serviceId"]
			if !ok {
				serviceId = "unknown"
			}
			if id, ok := serviceId.(string); ok {
				serviceId = id
			} else {
				serviceId = "unknown"
			}
			return nil, &customerrors.HTTPError{
				Status: 404,
				Msg:    fmt.Sprintf("Service not found: %s", serviceId),
			}
		}

		// Find all services that depend on the service with the given ID

		result, err = tx.Run(ctx, query, parameters)
		if err != nil {
			return nil, err
		}

		var dependencies = []*repositories.Dependency{}
		records, err = result.Collect(ctx)
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			id, _ := record.Get("id")
			name, _ := record.Get("name")
			version, _ := record.Get("version")
			interactionType, _ := record.Get("interaction_type")
			serviceType, _ := record.Get("type")
			dependency := &repositories.Dependency{
				Id: id.(string),
			}

			// Only set name and version if they exist
			if name != nil {
				dependency.Name = name.(string)
			}
			if version != nil {
				dependency.Version = version.(string)
			}
			if interactionType != nil {
				dependency.InteractionType = interactionType.(string)
			}
			if dependency.InteractionType == "" {
				dependency.InteractionType = "data"
			}

			if serviceType != nil {
				dependency.ServiceType = serviceType.(string)
			}

			dependencies = append(dependencies, dependency)
		}

		return dependencies, nil
	}
}
