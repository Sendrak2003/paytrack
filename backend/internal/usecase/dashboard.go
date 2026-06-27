// Package usecase содержит application-level бизнес-логику.
// Зависит только от domain (через интерфейсы), не от инфраструктуры.
package usecase

import "payments-dashboard/internal/domain"

// DashboardUseCase — use case для получения сводки и проектов.
type DashboardUseCase struct {
	repo domain.Repository
}

func NewDashboardUseCase(repo domain.Repository) *DashboardUseCase {
	return &DashboardUseCase{repo: repo}
}

func (uc *DashboardUseCase) GetSummary() (*domain.DashboardSummary, error) {
	return uc.repo.GetDashboardSummary()
}

func (uc *DashboardUseCase) GetProjects() ([]domain.ProjectSummary, error) {
	return uc.repo.GetProjectSummaries()
}

func (uc *DashboardUseCase) GetClients() ([]domain.Client, error) {
	return uc.repo.GetAllClients()
}
