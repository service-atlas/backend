package servicerepository

import (
	"context"
	"errors"
	"service-atlas/internal/auth"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (d *Neo4jServiceRepository) CreateService(ctx context.Context, service repositories.Service) (id string, err error) {
	createdBy := auth.NameFromContext(ctx)
	createServiceTransaction := func(tx neo4j.ManagedTransaction) (any, error) {
		props := map[string]any{
			"name":              service.Name,
			"type":              strings.ToUpper(service.ServiceType),
			"description":       service.Description,
			"url":               service.Url,
			"tier":              service.Tier,
			"architecture_role": service.ArchitectureRole,
			"exposure":          service.Exposure,
			"impact_domain":     service.ImpactDomain,
		}
		if createdBy != "" {
			props["createdBy"] = createdBy
		}

		result, err := tx.Run(
			ctx, `
        CREATE (n: Service)
        SET n = $props
        SET n.id = randomuuid(), n.created = datetime()
        RETURN n.id AS id
        `, map[string]any{
				"props": props,
			})
		if err != nil {
			return nil, err
		}
		svc, err := result.Single(ctx)
		if err != nil {
			return nil, err
		}
		svcMap := svc.AsMap()
		svcId, ok := svcMap["id"]
		if !ok {
			return nil, errors.New("failed to extract id from created service: id key missing")
		}
		idStr, ok := svcId.(string)
		if !ok {
			return nil, errors.New("failed to extract id from created service: id is not a string")
		}
		return idStr, nil

	}
	newId, insertErr := d.manager.ExecuteWrite(ctx, createServiceTransaction)
	if insertErr != nil {
		return "", insertErr
	}
	return newId.(string), nil
}

func (d *Neo4jServiceRepository) UpdateService(ctx context.Context, service repositories.Service) (err error) {
	updatedBy := auth.NameFromContext(ctx)
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
		props := map[string]any{
			"name":              service.Name,
			"type":              strings.ToUpper(service.ServiceType),
			"description":       service.Description,
			"url":               service.Url,
			"tier":              service.Tier,
			"architecture_role": service.ArchitectureRole,
			"exposure":          service.Exposure,
			"impact_domain":     service.ImpactDomain,
		}
		if updatedBy != "" {
			props["updatedBy"] = updatedBy
		}

		updateResult, updateErr := tx.Run(ctx, `
			MATCH (s:Service)
			WHERE s.id = $id
			SET s += $props
			SET s.updated = datetime()
			RETURN s
		`, map[string]any{
			"id":    service.Id,
			"props": props,
		})

		if updateErr != nil {
			return nil, updateErr
		}

		// Confirm update was successful
		if !updateResult.Next(ctx) {
			return nil, errors.New("update Service failed")
		}

		return nil, nil
	}

	_, execErr := d.manager.ExecuteWrite(ctx, updateServiceTransaction)
	if execErr != nil {
		return execErr
	}

	return nil
}
