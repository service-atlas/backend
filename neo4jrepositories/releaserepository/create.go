package releaserepository

import (
	"context"
	"fmt"
	"service-atlas/internal/auth"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r *Neo4jReleaseRepository) CreateRelease(ctx context.Context, release repositories.Release) error {
	createdBy := auth.NameFromContext(ctx)
	createReleaseTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		// Check if the service exists
		checkQuery := `
			MATCH (s:Service {id: $serviceId})
			RETURN s
		`
		result, err := tx.Run(ctx, checkQuery, map[string]any{
			"serviceId": release.ServiceId,
		})
		if err != nil {
			return nil, err
		}

		// If no records are returned, the service doesn't exist
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, &customerrors.HTTPError{
				Status: 404,
				Msg:    fmt.Sprintf("Service not found: %s", release.ServiceId),
			}
		}

		props := map[string]any{
			"releaseDate": release.ReleaseDate,
		}
		if release.Url != "" {
			props["url"] = release.Url
		}
		if release.Version != "" {
			props["version"] = release.Version
		}
		if createdBy != "" {
			props["createdBy"] = createdBy
		}

		query := `
			MATCH (s:Service {id: $serviceId})
			CREATE (r:Release)
			SET r = $props
			CREATE (s)-[rel:RELEASED]->(r)
			RETURN r
		`

		params := map[string]any{
			"serviceId": release.ServiceId,
			"props":     props,
		}

		_, err = tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, err := r.manager.ExecuteWrite(ctx, createReleaseTransaction)
	return err
}
