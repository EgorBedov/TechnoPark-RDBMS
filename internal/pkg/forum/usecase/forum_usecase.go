package usecase

import (
	"egogoger/internal/pkg/forum"
	"fmt"
)

type forumUseCase struct {
	forumRepo		forum.Repository
}

func NewForumUseCase(f forum.Repository) forum.UseCase {
	return &forumUseCase{forumRepo: f}
}

func (fu *forumUseCase) Echo() {
	fmt.Println("Forum usecase")

	fu.forumRepo.Echo()
}
