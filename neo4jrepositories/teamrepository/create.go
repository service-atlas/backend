package teamrepository

import (
	"context"
	"net/http"
	"service-atlas/internal/auth"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Neo4jTeamRepository) CreateTeam(ctx context.Context, team repositories.Team) (string, error) {
	createdBy := auth.NameFromContext(ctx)
	createTeamTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		props := map[string]any{
			"name": team.Name,
		}
		if createdBy != "" {
			props["createdBy"] = createdBy
		}

		result, err := tx.Run(
			ctx, `
        CREATE (n: Team)
        SET n = $props
        SET n.id = randomuuid(), n.created = datetime(), n.updated = datetime()
        RETURN n.id AS id
        `, map[string]any{
				"props": props,
			})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			id, ok := result.Record().Get("id")
			if !ok {
				return nil, &customerrors.HTTPError{
					Status: http.StatusInternalServerError,
					Msg:    "Id not returned when creating team",
				}
			}
			return id, nil
		}
		return nil, &customerrors.HTTPError{
			Status: http.StatusInternalServerError,
			Msg:    "No id returned from creating team",
		}

	}
	result, err := r.manager.ExecuteWrite(ctx, createTeamTransaction)
	if err != nil {
		return "", &customerrors.HTTPError{
			Status: http.StatusInternalServerError,
			Msg:    "Error creating team",
		}
	}
	id, ok := result.(string)
	if !ok {
		return "", &customerrors.HTTPError{
			Status: http.StatusInternalServerError,
			Msg:    "Error creating team",
		}
	}
	return id, nil
}
