package thread

import "egogoger/internal/pkg/models"

type Repository interface {
	CreatePosts([]models.Post, int) int
	GetInfo(*models.Thread, string) int
	UpdateThread(*models.Thread, string) int
	GetPosts(*models.PostQuery) ([]models.Post, int)
	Vote(*models.Vote) (int, int)

	// Private
	GetThreadIdBySlugOrId(string) (int, error)
}
