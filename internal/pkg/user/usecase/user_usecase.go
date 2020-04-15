package usecase

import (
	"egogoger/internal/pkg/user"
	"fmt"
)

type userUseCase struct {
	userRepo		user.Repository
}

func NewUserUseCase(f user.Repository) user.UseCase {
	return &userUseCase{userRepo: f}
}

func (fu *userUseCase) Echo() {
	fmt.Println("User usecase")

	fu.userRepo.Echo()
}
