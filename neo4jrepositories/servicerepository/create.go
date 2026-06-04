package servicerepository

import (
	"context"
	"service-atlas/internal/auth"
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
			return "", err
		}
		svc, err := result.Single(ctx)
		if err != nil {
			return "", err
		}
		svcMap := svc.AsMap()
		if svcId, ok := svcMap["id"]; ok {
			if idStr, ok := svcId.(string); ok {
				return idStr, err
			}
		}
		return "", err

	}
	newId, insertErr := d.manager.ExecuteWrite(ctx, createServiceTransaction)
	if insertErr != nil {
		return "", insertErr
	}
	return newId.(string), nil
}
