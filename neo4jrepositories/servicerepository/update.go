package servicerepository

import (
	"context"
	"errors"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (d *Neo4jServiceRepository) UpdateService(ctx context.Context, service repositories.Service) (err error) {
	updateServiceTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		// First check if the service exists
		result, err := tx.Run(ctx, `
			MATCH (s:Service)
			WHERE s.id = $id
			RETURN s
		`, map[string]any{
			"id": service.Id,
		})

		if err != nil {
			return nil, err
		}

		found := result.Next(ctx)
		if !found {
			return nil, &customerrors.HTTPError{
				Status: 404,
				Msg:    "Service not found",
			}
		}
		// Service exists, update it
		updateResult, updateErr := tx.Run(ctx, `
			MATCH (s:Service)
			WHERE s.id = $id
			SET s.name = $name, 
				s.type = $type, 
				s.description = $description,
				s.url = $url,
				s.criticality = $criticality,
				s.updated = datetime()
			RETURN s
		`, map[string]any{
			"id":          service.Id,
			"name":        service.Name,
			"type":        service.ServiceType,
			"description": service.Description,
			"url":         service.Url,
			"criticality": service.Criticality,
		})

		if updateErr != nil {
			err = updateErr
		}

		// Confirm update was successful
		if !updateResult.Next(ctx) {
			err = errors.New("update Service failed")
		}

		return nil, err
	}

	_, execErr := d.manager.ExecuteWrite(ctx, updateServiceTransaction)
	if execErr != nil {
		return execErr
	}

	return nil
}
