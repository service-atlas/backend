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
		return "", err
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

func (r Neo4jTeamRepository) UpdateTeam(ctx context.Context, team repositories.Team) error {
	updatedBy := auth.NameFromContext(ctx)
	_, err := r.GetTeam(ctx, team.Id)
	// Error should be a custom http error already
	if err != nil {
		return err
	}

	updateTeamTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		props := map[string]any{
			"name": team.Name,
		}
		if updatedBy != "" {
			props["updatedBy"] = updatedBy
		}

		updateResult, updateErr := tx.Run(ctx, `
			MATCH (s:Team)
			WHERE s.id = $id
			SET s += $props,
				s.updated = datetime()
			RETURN s
		`, map[string]any{
			"id":    team.Id,
			"props": props,
		})

		if updateErr != nil {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    updateErr.Error(),
			}
		}

		// Confirm update was successful
		if !updateResult.Next(ctx) {
			if resultErr := updateResult.Err(); resultErr != nil {
				return nil, &customerrors.HTTPError{
					Status: http.StatusInternalServerError,
					Msg:    resultErr.Error(),
				}
			}
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to confirm update",
			}
		}
		return nil, nil
	}

	_, err = r.manager.ExecuteWrite(ctx, updateTeamTransaction)
	if err != nil {
		return err
	}
	return nil
}
