package repositories

import (
	"context"
	"time"
)

// DebtRepository defines the methods for interacting with debt items.
type DebtRepository interface {
	// CreateDebtItem creates a new debt item.
	CreateDebtItem(ctx context.Context, debt Debt) error
	// UpdateStatus updates the status of an existing debt item.
	UpdateStatus(ctx context.Context, id, status string) error
	// GetDebtByServiceId retrieves debt items for a given service.
	GetDebtByServiceId(ctx context.Context, id string, page, pageSize int, onlyResolved bool) ([]Debt, error)
}

// ServiceRepository defines the methods for interacting with services.
type ServiceRepository interface {
	// GetAllServices retrieves all services.
	GetAllServices(ctx context.Context, page int, pageSize int) ([]Service, error)
	// CreateService creates a new service.
	CreateService(ctx context.Context, service Service) (string, error)
	// UpdateService updates an existing service.
	UpdateService(ctx context.Context, service Service) error
	// DeleteService deletes a service.
	DeleteService(ctx context.Context, id string) error
	// GetServiceById retrieves a service by its ID.
	GetServiceById(ctx context.Context, id string) (Service, error)
	// Search performs a fuzzy search against the Service full-text index and returns matching services ordered by relevance.
	Search(ctx context.Context, query string) ([]Service, error)
	// GetTeamsByServiceId retrieves all teams associated with a service.
	GetTeamsByServiceId(ctx context.Context, serviceId string) ([]Team, error)
}

// DependencyRepository defines the methods for interacting with dependencies.
type DependencyRepository interface {
	// AddDependency adds a dependency to a resource.
	AddDependency(ctx context.Context, id string, dependency Dependency) error
	// GetDependencies retrieves all dependencies of a resource.
	GetDependencies(ctx context.Context, id string) ([]*Dependency, error)
	// GetDependenciesByInteractionType retrieves all dependencies of a resource of a specific interaction type.
	GetDependenciesByInteractionType(ctx context.Context, id, interaction_type string) ([]*Dependency, error)
	// GetDependents retrieves all resources that depend on a given resource.
	GetDependents(ctx context.Context, id string) ([]*Dependency, error)
	// DeleteDependency deletes a dependency between two resources.
	DeleteDependency(ctx context.Context, id string, dependsOnID string) error
}

// ReleaseRepository defines the methods for interacting with releases.
type ReleaseRepository interface {
	// CreateRelease creates a new release.
	CreateRelease(ctx context.Context, release Release) error
	// GetReleasesByServiceId retrieves all releases associated with a service.
	GetReleasesByServiceId(ctx context.Context, serviceId string, page, pageSize int) ([]*Release, error)
	// GetReleasesInDateRange retrieves all releases within a specified date range.
	GetReleasesInDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*ServiceReleaseInfo, error)
}

// ReportRepository defines the methods for gathering reports.
type ReportRepository interface {
	// GetServiceRiskReport retrieves the risk report for a service.
	GetServiceRiskReport(ctx context.Context, serviceId string) (*ServiceRiskReport, error)
	// GetServiceChangeRisk computes the heuristic change risk for a service (low|medium|high).
	GetServiceChangeRisk(ctx context.Context, serviceId string) (*ServiceChangeRisk, error)
	// GetServicesByTeam retrieves all services associated with a team.
	GetServicesByTeam(ctx context.Context, teamId string) ([]Service, error)
	// GetDebtCountByService retrieves the number of debt items for each service.
	GetDebtCountByService(ctx context.Context) ([]ServiceDebtReport, error)
	// GetServicesByTier retrieves all services on a given criticality tier.
	GetServicesByTier(ctx context.Context, tier int) ([]Service, error)
	// GetServiceTypes retrieves all service types and counts for each type.
	GetServiceTypes(ctx context.Context) ([]ServiceType, error)
}

// TeamRepository defines the methods for interacting with teams.
type TeamRepository interface {
	// CreateTeam creates a new team.
	CreateTeam(ctx context.Context, team Team) (string, error)
	// GetTeam retrieves a team by its ID.
	GetTeam(ctx context.Context, teamId string) (*Team, error)
	// GetTeams retrieves all teams.
	GetTeams(ctx context.Context, page, pageSize int) ([]Team, error)
	// UpdateTeam updates an existing team.
	UpdateTeam(ctx context.Context, team Team) error
	// DeleteTeam deletes a team.
	DeleteTeam(ctx context.Context, teamId string) error
	// CreateTeamAssociation creates a new team association with a service.
	CreateTeamAssociation(ctx context.Context, teamId, serviceId string) error
	// DeleteTeamAssociation deletes a team association with a service.
	DeleteTeamAssociation(ctx context.Context, teamId, serviceId string) error
}
