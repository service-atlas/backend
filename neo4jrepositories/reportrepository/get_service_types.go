package reportrepository

import (
	"context"
	"errors"
	"service-atlas/internal"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Neo4jReportRepository) GetServiceTypes(ctx context.Context) ([]repositories.ServiceType, error) {
	typeReport := make([]repositories.ServiceType, 2)
	_, err := r.manager.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		typeCountQuery := `
			MATCH (s:Service)
			RETURN s.type AS type, count(*) AS count
			ORDER BY count DESC
			`
		result, err := tx.Run(ctx, typeCountQuery, nil)

		if err != nil {
			return nil, err
		}
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			t, ok := record.Get("type")
			if !ok {
				return nil, errors.New("type not found")
			}
			serviceType := repositories.ServiceType{}
			serviceType.Type = internal.ToTitleCase(t.(string))
			count, ok := record.Get("count")
			if !ok {
				return nil, errors.New("count not found")
			}
			serviceType.Count = count.(int64)
			typeReport = append(typeReport, serviceType)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	return typeReport, nil
}
