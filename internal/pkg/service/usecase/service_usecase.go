package usecase

import (
	"egogoger/internal/pkg/cache"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/service"
)

type serviceUseCase struct {
	serviceRepo		service.Repository
}

func NewServiceUseCase(f service.Repository) service.UseCase {
	return &serviceUseCase{serviceRepo: f}
}

func (su *serviceUseCase) TruncateAll() int {
	cache.ClearThreadsCache()
	cache.ClearForumsCache()
	return su.serviceRepo.TruncateAll()
}

func (su *serviceUseCase) GetInfo() (*models.Summary, int) {
	return su.serviceRepo.GetInfo()
}
