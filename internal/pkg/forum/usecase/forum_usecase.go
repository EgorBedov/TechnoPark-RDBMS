package usecase

import (
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
)

type forumUseCase struct {
	forumRepo		forum.Repository
}

func NewForumUseCase(f forum.Repository) forum.UseCase {
	return &forumUseCase{forumRepo: f}
}

func (fu *forumUseCase) CreateForum(frm *models.Forum) int {
	return fu.forumRepo.CreateForum(frm)
}

func (fu *forumUseCase) GetInfo(frm *models.Forum) int {
	return fu.forumRepo.GetInfo(frm)
}

func (fu *forumUseCase) CreateThread(thrd *models.Thread) int {
	return fu.forumRepo.CreateThread(thrd)
}

func (fu *forumUseCase) GetUsers(query models.Query) ([]models.User, int) {
	return fu.forumRepo.GetUsers(query)
}

func (fu *forumUseCase) GetThreads(query models.Query) ([]models.Thread, int) {
	return fu.forumRepo.GetThreads(query)
}
