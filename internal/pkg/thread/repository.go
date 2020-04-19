package thread

import "egogoger/internal/pkg/models"

type Repository interface {
	CreatePosts([]models.Post, string) int
	GetInfo(*models.Thread, string) int
	UpdateThread(*models.Thread, string) int
	GetPosts(*models.PostQuery) ([]models.Post, int)
	Vote(*models.Vote) int
}
