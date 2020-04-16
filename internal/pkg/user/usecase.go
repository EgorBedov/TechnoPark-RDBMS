package user

import (
	"egogoger/internal/pkg/models"
)

type UseCase interface {
	CreateUser(*models.User) int
	GetInfo(*models.User) int
	PostInfo(*models.User) int
}