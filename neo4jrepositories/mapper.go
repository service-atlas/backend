package neo4jrepositories

import (
	"service-atlas/internal"
	"service-atlas/repositories"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// MapNodeToService converts a Neo4j node to a Service object
func MapNodeToService(n neo4j.Node) repositories.Service {
	svc := repositories.Service{}

	// Safely extract name with validation
	if name, ok := n.Props["name"]; ok {
		if nameStr, ok := name.(string); ok {
			svc.Name = nameStr
		}
	}

	// Safely extract description with validation
	if desc, ok := n.Props["description"]; ok {
		if descStr, ok := desc.(string); ok {
			svc.Description = descStr
		}
	}

	// Safely extract service type with validation
	if svcType, ok := n.Props["type"]; ok {
		if typeStr, ok := svcType.(string); ok {
			svc.ServiceType = internal.ToTitleCase(typeStr)
		}
	}

	// Safely extract ID with validation
	if id, ok := n.Props["id"]; ok {
		if idStr, ok := id.(string); ok {
			svc.Id = idStr
		}
	}

	if url, ok := n.Props["url"]; ok {
		if urlStr, ok := url.(string); ok {
			svc.Url = urlStr
		}
	}

	if crit, ok := n.Props["tier"]; ok {
		if critInt64, ok := crit.(int64); ok {
			svc.Tier = int(critInt64)
		} else if critInt, ok := crit.(int); ok {
			svc.Tier = critInt
		}
	}

	// Safely extract created date with validation
	if date, ok := n.Props["created"]; ok {
		if dateStr, ok := date.(time.Time); ok {
			svc.Created = dateStr
		}
	}

	if date, ok := n.Props["updated"]; ok {
		if dateStr, ok := date.(time.Time); ok {
			svc.Updated = dateStr
		}
	}

	if role, ok := n.Props["architecture_role"]; ok {
		if roleStr, ok := role.(string); ok {
			svc.ArchitectureRole = roleStr
		}
	}

	if exposure, ok := n.Props["exposure"]; ok {
		if exposureStr, ok := exposure.(string); ok {
			svc.Exposure = exposureStr
		}
	}

	if domain, ok := n.Props["impact_domain"]; ok {
		if domainList, ok := domain.([]any); ok {
			for _, d := range domainList {
				if dStr, ok := d.(string); ok {
					svc.ImpactDomain = append(svc.ImpactDomain, dStr)
				}
			}
		}
	}
	return svc
}

// MapNodeToTeam converts a Neo4j node to a Team object
func MapNodeToTeam(n neo4j.Node) (repositories.Team, bool) {
	team := repositories.Team{}

	// Safely extract name with validation
	if name, ok := getPropFromNode[string](n, "name"); ok {
		team.Name = name
	} else {
		return team, false
	}

	// Safely extract ID with validation
	if id, ok := getPropFromNode[string](n, "id"); ok {
		team.Id = id
	} else {
		return team, false
	}

	// Safely extract created date with validation
	if date, ok := getPropFromNode[time.Time](n, "created"); ok {
		team.Created = date
	} else {
		return team, false
	}
	if date, ok := getPropFromNode[time.Time](n, "updated"); ok {
		team.Updated = date
	} else {
		return team, false
	}

	return team, true
}

func getPropFromNode[T string | time.Time](n neo4j.Node, key string) (T, bool) {
	if value, ok := n.Props[key]; ok {
		if v, ok := value.(T); ok {
			return v, true
		}
		return *new(T), false
	}
	return *new(T), false
}
