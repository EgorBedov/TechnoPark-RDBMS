package thread

import "egogoger/internal/pkg/models"

type UseCase interface {
	CreatePosts([]models.Post, string) int
	GetInfo()
	PostInfo()
	GetPosts()
	Vote()
}