package reportrepository

import (
	"context"
	"service-atlas/internal"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Neo4jReportRepository) GetServicesByTier(ctx context.Context, tier int) ([]repositories.Service, error) {
	services := make([]repositories.Service, 0)
	logger := internal.LoggerFromContext(ctx)
	logger.Info("Getting services by tier", "tier", tier)
	work := func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (s:Service)
			WHERE s.tier = $tier
			RETURN s
		`, map[string]any{
			"tier": tier,
		})
		if err != nil {
			return nil, err
		}
		for result.Next(ctx) {
			record := result.Record()
			node, ok := record.Get("s")
			if !ok {
				logger.Warn("Failed to extract node from query result", "record", record)
				continue
			}

			n, ok := node.(neo4j.Node)
			if !ok {
				logger.Warn("Failed to convert node to Service", "node", node)
				continue
			}

			svc := nRepo.MapNodeToService(n)
			services = append(services, svc)
		}
		if err := result.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	_, err := r.manager.ExecuteRead(ctx, work)
	if err != nil {
		return nil, err
	}
	return services, nil

}
