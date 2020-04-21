package usecase

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/user"
)

type userUseCase struct {
	userRepo		user.Repository
}

func NewUserUseCase(f user.Repository) user.UseCase {
	return &userUseCase{userRepo: f}
}

func (uu *userUseCase) CreateUser(usr *models.User) ([]models.User, int) {
	return uu.userRepo.CreateUser(usr)
}

func (uu *userUseCase) GetInfo(user *models.User) int {
	return uu.userRepo.GetInfo(user)
}

func (uu *userUseCase) PostInfo(user *models.User) (int, *models.Message) {
	return uu.userRepo.PostInfo(user)
}
