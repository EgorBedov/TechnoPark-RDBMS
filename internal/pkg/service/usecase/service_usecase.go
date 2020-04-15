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

func (fu *serviceUseCase) Echo() {
	fmt.Println("Service usecase")

	fu.serviceRepo.Echo()
}
