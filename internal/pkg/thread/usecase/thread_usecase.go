package usecase

import (
	"egogoger/internal/pkg/thread"
	"fmt"
)

type threadUseCase struct {
	threadRepo		thread.Repository
}

func NewThreadUseCase(f thread.Repository) thread.UseCase {
	return &threadUseCase{threadRepo: f}
}

func (tu *threadUseCase) CreatePosts() {
	fmt.Println("Thread usecase CreatePosts")
	tu.threadRepo.CreatePosts()
}

func (tu *threadUseCase) GetInfo() {
	fmt.Println("Thread usecase GetInfo")
	tu.threadRepo.GetInfo()
}

func (tu *threadUseCase) PostInfo() {
	fmt.Println("Thread usecase PostInfo")
	tu.threadRepo.PostInfo()
}

func (tu *threadUseCase) GetPosts() {
	fmt.Println("Thread usecase GetPosts")
	tu.threadRepo.GetPosts()
}

func (tu *threadUseCase) Vote() {
	fmt.Println("Thread usecase Vote")
	tu.threadRepo.Vote()
}
