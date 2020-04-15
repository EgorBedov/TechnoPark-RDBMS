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

func (fu *threadUseCase) Echo() {
	fmt.Println("Thread usecase")

	fu.threadRepo.Echo()
}
