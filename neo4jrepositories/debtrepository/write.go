package debtrepository

import (
	"context"
	"fmt"
	"service-atlas/internal/auth"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (n Neo4jDebtRepository) CreateDebtItem(ctx context.Context, debt repositories.Debt) error {
	createdBy := auth.NameFromContext(ctx)
	createDebtTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		// Check if the service exists
		checkQuery := `
			MATCH (s:Service {id: $serviceId})
			RETURN s
		`
		result, err := tx.Run(ctx, checkQuery, map[string]any{
			"serviceId": debt.ServiceId,
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
				Msg:    fmt.Sprintf("Service not found: %s", debt.ServiceId),
			}
		}

		cypher := `
				MATCH (s:Service {id: $serviceId})
				CREATE (n:Debt)
				SET n = $props
				SET n.id = randomuuid(), n.created = datetime()
				CREATE (s)-[:OWNS]->(n)
        `
		props := map[string]any{
			"title":       debt.Title,
			"type":        debt.Type,
			"description": debt.Description,
			"status":      DefaultStatus,
		}

		if createdBy != "" {
			props["createdBy"] = createdBy
		}

		params := map[string]any{
			"serviceId": debt.ServiceId,
			"props":     props,
		}

		_, err = tx.Run(ctx, cypher, params)
		return nil, err
	}
	_, err := n.manager.ExecuteWrite(ctx, createDebtTransaction)
	return err
}

func (n Neo4jDebtRepository) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := n.manager.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (d:Debt {id: $id})
			SET d.status = $status
			RETURN count(d) as updatedCount
		`, map[string]any{
			"id":     id,
			"status": status,
		})
		if err != nil {
			return nil, err
		}
		if result.Next(ctx) {
			ctr, _ := result.Record().Get("updatedCount")
			if ct, ok := ctr.(int64); ok && ct == 0 {
				return nil, &customerrors.HTTPError{
					Status: 404,
					Msg:    "Debt not found",
				}
			}
		} else if result.Err() != nil {
			return nil, result.Err()
		}
		return nil, nil
	})
	return err
}
