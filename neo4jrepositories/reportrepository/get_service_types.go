package reportrepository

import (
	"context"
	"errors"
	"service-atlas/internal"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func (r Neo4jReportRepository) GetServiceTypes(ctx context.Context) ([]repositories.ServiceType, error) {

	result, err := r.manager.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
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
		var localReport []repositories.ServiceType
		for _, record := range records {
			t, ok := record.Get("type")
			if !ok {
				return nil, errors.New("type not found")
			}
			serviceType := repositories.ServiceType{}
			if t == nil {
				return nil, errors.New("type value is nil")
			}
			tVal, ok := t.(string)
			if !ok {
				return nil, errors.New("type value is not a string")
			}
			serviceType.Type = internal.ToTitleCase(tVal)

			count, ok := record.Get("count")
			if !ok {
				return nil, errors.New("count not found")
			}
			if count == nil {
				return nil, errors.New("count value is nil")
			}
			countVal, ok := count.(int64)
			if !ok {
				return nil, errors.New("count value is not an int64")
			}
			serviceType.Count = countVal
			localReport = append(localReport, serviceType)
		}

		return localReport, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]repositories.ServiceType), nil
}
