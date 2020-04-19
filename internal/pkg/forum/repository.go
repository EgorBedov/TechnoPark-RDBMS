package forum

import "egogoger/internal/pkg/models"

type Repository interface {
	CreateForum(*models.Forum) int
	GetInfo(*models.Forum) int
	CreateThread(*models.Thread) int
	GetUsers(models.Query) ([]models.User, int)
	GetThreads(models.Query) ([]models.Thread, int)
}
