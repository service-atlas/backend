package reports

import (
	"context"
	"service-atlas/repositories"
)

// mockReportRepository is a mock implementation of the ReportRepository interface
type mockReportRepository struct {
	Err      error
	Report   *repositories.ServiceRiskReport
	Services []repositories.Service
	Debt     []repositories.ServiceDebtReport
}

func (repo mockReportRepository) GetServiceRiskReport(_ context.Context, _ string) (*repositories.ServiceRiskReport, error) {
	if repo.Err != nil {
		return nil, repo.Err
	}
	return repo.Report, nil
}

func (repo mockReportRepository) GetServicesByTeam(_ context.Context, _ string) ([]repositories.Service, error) {
	if repo.Err != nil {
		return nil, repo.Err
	}
	if repo.Services != nil {
		return repo.Services, nil
	}
	return []repositories.Service{}, nil
}

func (repo mockReportRepository) GetDebtCountByService(_ context.Context) ([]repositories.ServiceDebtReport, error) {
	if repo.Err != nil {
		return nil, repo.Err
	}
	return repo.Debt, nil
}

func (repo mockReportRepository) GetServicesByTier(_ context.Context, _ int) ([]repositories.Service, error) {
	if repo.Err != nil {
		return nil, repo.Err
	}
	return repo.Services, nil
}
