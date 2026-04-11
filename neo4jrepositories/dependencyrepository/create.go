package dependencyrepository

import (
	"context"
	"fmt"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (d *Neo4jDependencyRepository) AddDependency(ctx context.Context, id string, dependency repositories.Dependency) error {
	createDependencyTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		// Check if both services exist
		checkQuery := `
			MATCH (s1:Service {id: $serviceId})
			MATCH (s2:Service {id: $dependencyId})
			RETURN s1, s2
		`
		result, err := tx.Run(ctx, checkQuery, map[string]any{
			"serviceId":    id,
			"dependencyId": dependency.Id,
		})
		if err != nil {
			return nil, err
		}

		// If no records are returned, one or both services don't exist
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, &customerrors.HTTPError{
				Status: 404,
				Msg:    fmt.Sprintf("One or both services not found: %s, %s", id, dependency.Id),
			}
		}

		// Create the dependency relationship
		query := `
			MATCH (s1:Service {id: $serviceId})
			MATCH (s2:Service {id: $dependencyId})
			MERGE (s1)-[r:DEPENDS_ON]->(s2)
			SET r.interaction_type = $interactionType
			WITH r, 
				 CASE WHEN $version IS NOT NULL AND $version <> "" 
					  THEN $version 
					  ELSE r.version 
				 END AS resolvedVersion
			SET r.version = resolvedVersion
			RETURN r
		`

		params := map[string]any{
			"serviceId":       id,
			"dependencyId":    dependency.Id,
			"version":         dependency.Version,
			"interactionType": dependency.InteractionType,
		}

		_, err = tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, err := d.manager.ExecuteWrite(ctx, createDependencyTransaction)
	return err
}
