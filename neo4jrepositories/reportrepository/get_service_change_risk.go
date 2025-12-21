package reportrepository

import (
	"context"
	"fmt"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// GetServiceChangeRisk computes a heuristic change risk for a service based on:
// - Service tier (base score)
// - Incoming dependents up to depth 2 (depth 1 at 100%, depth 2 at 50%)
// The final qualitative risk buckets are:
//
//	0–29 => low, 30–69 => medium, 70+ => high
func (r Neo4jReportRepository) GetServiceChangeRisk(ctx context.Context, serviceId string) (*repositories.ServiceChangeRisk, error) {
	// First, fetch the service tier; return 404 if not found
	var svcTier int64
	_, err := r.manager.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
            MATCH (s:Service {id: $serviceId})
            RETURN s.tier AS tier
        `, map[string]any{"serviceId": serviceId})
		if err != nil {
			return nil, err
		}
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return nil, &customerrors.HTTPError{Status: 404, Msg: fmt.Sprintf("Service not found: %s", serviceId)}
		}
		val, _ := records[0].Get("tier")
		if t, ok := val.(int64); ok {
			svcTier = t
		} else {
			// default tier 3 if missing or not an int
			svcTier = 3
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	base := baseScoreByTier(int(svcTier))

	// Next, gather incoming dependents up to depth 2 and compute weighted sum
	var depScore float64
	_, err = r.manager.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// For each dependent, take the minimal path length to the target (1 or 2)
		cypher := `
            MATCH p = (dep:Service)-[:DEPENDS_ON*1..2]->(s:Service {id: $serviceId})
            WITH dep, min(length(p)) AS depth
            RETURN depth AS depth, dep.tier AS tier
        `
		result, err := tx.Run(ctx, cypher, map[string]any{"serviceId": serviceId})
		if err != nil {
			return nil, err
		}
		for result.Next(ctx) {
			rec := result.Record().AsMap()
			dRaw, ok := rec["depth"].(int64)
			if !ok {
				continue
			}
			tRaw, _ := rec["tier"].(int64)
			add := dependentScoreByTier(int(tRaw))
			switch d := int(dRaw); d {
			case 1:
				depScore += float64(add)
			case 2:
				depScore += float64(add) * 0.5
			default:
				// ignore depth >= 3
			}
		}
		if err := result.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	total := int(float64(base) + depScore)
	risk := bucketizeChangeRisk(total)

	return &repositories.ServiceChangeRisk{Risk: risk, Score: total}, nil
}

func baseScoreByTier(tier int) int {
	switch tier {
	case 1:
		return 40
	case 2:
		return 25
	case 3:
		return 10
	case 4:
		return 5
	default:
		return 10 // default similar to tier 3 if unset
	}
}

func dependentScoreByTier(tier int) int {
	switch tier {
	case 1:
		return 20
	case 2:
		return 12
	case 3:
		return 5
	case 4:
		return 2
	default:
		return 0
	}
}

func bucketizeChangeRisk(score int) string {
	switch {
	case score >= 70:
		return "high"
	case score >= 30:
		return "medium"
	default:
		return "low"
	}
}
