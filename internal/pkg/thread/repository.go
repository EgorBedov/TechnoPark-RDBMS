package thread

import "egogoger/internal/pkg/models"

type Repository interface {
	CreatePosts([]models.Post, string) int
	GetInfo()
	PostInfo()
	GetPosts()
	Vote()
}
