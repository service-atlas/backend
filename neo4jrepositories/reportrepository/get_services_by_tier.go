package reportrepository

import (
	"context"
	nRepo "service-atlas/neo4jrepositories"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Neo4jReportRepository) GetServicesByTier(ctx context.Context, tier int) ([]repositories.Service, error) {
	services := make([]repositories.Service, 0)
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
				continue
			}

			n, ok := node.(neo4j.Node)
			if !ok {
				continue
			}

			svc := nRepo.MapNodeToService(n)
			services = append(services, svc)
		}
		return nil, nil
	}
	_, err := r.manager.ExecuteRead(ctx, work)
	if err != nil {
		return nil, err
	}
	return services, nil

}
