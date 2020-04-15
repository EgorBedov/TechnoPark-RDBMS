package usecase

import (
	"egogoger/internal/pkg/service"
	"fmt"
)

type serviceUseCase struct {
	serviceRepo		service.Repository
}

func NewServiceUseCase(f service.Repository) service.UseCase {
	return &serviceUseCase{serviceRepo: f}
}

func (su *serviceUseCase) TruncateAll() {
	fmt.Println("Service usecase TruncateAll")
	su.serviceRepo.TruncateAll()
}

func (su *serviceUseCase) GetInfo() {
	fmt.Println("Service usecase GetInfo")
	su.serviceRepo.GetInfo()
}
