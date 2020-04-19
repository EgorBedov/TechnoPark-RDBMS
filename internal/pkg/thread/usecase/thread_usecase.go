package usecase

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/thread"
)

type threadUseCase struct {
	threadRepo		thread.Repository
}

func NewThreadUseCase(f thread.Repository) thread.UseCase {
	return &threadUseCase{threadRepo: f}
}

func (tu *threadUseCase) CreatePosts(posts []models.Post, slugOrId string) int {
	return tu.threadRepo.CreatePosts(posts, slugOrId)
}

func (tu *threadUseCase) GetInfo(thrd *models.Thread, slugOrId string) int {
	return tu.threadRepo.GetInfo(thrd, slugOrId)
}

func (tu *threadUseCase) UpdateThread(thrd *models.Thread, slugOrId string) int {
	return tu.threadRepo.UpdateThread(thrd, slugOrId)
}

func (tu *threadUseCase) GetPosts(query *models.PostQuery) ([]models.Post, int) {
	return tu.threadRepo.GetPosts(query)
}

func (tu *threadUseCase) Vote(vote *models.Vote) int {
	return tu.threadRepo.Vote(vote)
}
