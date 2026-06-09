package teamrepository

import (
	"context"
	"net/http"
	"service-atlas/internal/customerrors"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Neo4jTeamRepository) GetTeam(ctx context.Context, teamId string) (*repositories.Team, error) {
	getTeamTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (s:Team)
			WHERE s.id = $id
			RETURN s
		`, map[string]any{
			"id": teamId,
		})

		if err != nil {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    err.Error(),
			}
		}

		if !result.Next(ctx) {
			if err := result.Err(); err != nil {
				return nil, &customerrors.HTTPError{
					Status: http.StatusInternalServerError,
					Msg:    err.Error(),
				}
			}
			return nil, &customerrors.HTTPError{
				Status: http.StatusNotFound,
				Msg:    "Team not found",
			} // No team found with this ID
		}

		record := result.Record()
		node, ok := record.Get("s")
		if !ok {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to extract team node from query result",
			}
		}

		n, ok := node.(neo4j.Node)
		if !ok {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to convert query result to Node type",
			}
		}
		t, ok := nRepo.MapNodeToTeam(n)
		if !ok {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    "Failed to convert Node to Team type",
			}
		}
		return t, nil
	}
	result, err := r.manager.ExecuteRead(ctx, getTeamTransaction)
	if err != nil {
		return nil, err
	}
	team, ok := result.(repositories.Team)
	if !ok {
		return nil, &customerrors.HTTPError{
			Status: http.StatusInternalServerError,
			Msg:    "Error converting result to team",
		}
	}
	return &team, nil

}

func (r Neo4jTeamRepository) GetTeams(ctx context.Context, page, pageSize int) ([]repositories.Team, error) {
	getPageTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		skip := (page - 1) * pageSize

		result, err := tx.Run(ctx, `
		    MATCH (s:Team)
			RETURN s
			ORDER BY s.created DESC
			SKIP $skip
			LIMIT $limit
		`, map[string]any{
			"skip":  skip,
			"limit": pageSize,
		})
		if err != nil {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    err.Error(),
			}
		}
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, &customerrors.HTTPError{
				Status: http.StatusInternalServerError,
				Msg:    err.Error(),
			}
		}
		teams := []repositories.Team{}
		for _, record := range records {
			node, ok := record.Get("s")
			if !ok {
				continue
			}
			n, ok := node.(neo4j.Node)
			if !ok {
				continue
			}
			team, ok := nRepo.MapNodeToTeam(n)
			if !ok {
				continue
			}
			teams = append(teams, team)
		}
		return teams, nil
	}
	pagedResult, err := r.manager.ExecuteRead(ctx, getPageTransaction)
	if err != nil {
		return nil, &customerrors.HTTPError{
			Status: http.StatusInternalServerError,
			Msg:    err.Error(),
		}
	}
	teams, ok := pagedResult.([]repositories.Team)
	if !ok {
		return nil, &customerrors.HTTPError{
			Status: http.StatusInternalServerError,
			Msg:    "unexpected return type from transaction",
		}
	}
	return teams, nil
}
