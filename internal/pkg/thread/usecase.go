package thread

import "egogoger/internal/pkg/models"

type UseCase interface {
	CreatePosts([]models.Post, int, string) models.Message
	GetInfo(*models.Thread, string) int
	UpdateThread(*models.Thread, string) int
	GetPosts(*models.PostQuery) ([]models.Post, int)
	Vote(*models.Vote) (*models.Thread, models.Message)

	// Private
	GetThreadInfoBySlugOrId(string) (int, string, error)
}