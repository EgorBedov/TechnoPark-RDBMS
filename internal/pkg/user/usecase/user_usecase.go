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

func (uu *userUseCase) CreateUser() {
	fmt.Println("User usecase CreateUser")
	uu.userRepo.CreateUser()
}

func (uu *userUseCase) GetInfo() {
	fmt.Println("User usecase GetInfo")
	uu.userRepo.GetInfo()
}

func (uu *userUseCase) PostInfo() {
	fmt.Println("User usecase PostInfo")
	uu.userRepo.PostInfo()
}
