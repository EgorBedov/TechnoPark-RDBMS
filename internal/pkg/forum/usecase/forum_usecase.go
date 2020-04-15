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

func (fu *forumUseCase) CreateForum() {
	fmt.Println("Forum usecase CreateForum")
	fu.forumRepo.CreateForum()
}

func (fu *forumUseCase) GetInfo() {
	fmt.Println("Forum usecase GetInfo")
	fu.forumRepo.GetInfo()
}

func (fu *forumUseCase) CreateThread() {
	fmt.Println("Forum usecase CreateThread")
	fu.forumRepo.CreateThread()
}

func (fu *forumUseCase) GetUsers() {
	fmt.Println("Forum usecase GetUsers")
	fu.forumRepo.GetUsers()
}

func (fu *forumUseCase) GetThreads() {
	fmt.Println("Forum usecase GetThreads")
	fu.forumRepo.GetThreads()
}
