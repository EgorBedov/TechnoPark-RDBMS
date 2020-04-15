package usecase

import (
	"egogoger/internal/pkg/post"
	"fmt"
)

type postUseCase struct {
	postRepo		post.Repository
}

func NewPostUseCase(f post.Repository) post.UseCase {
	return &postUseCase{postRepo: f}
}

func (pu *postUseCase) GetInfo() {
	fmt.Println("Post usecase GetInfo")
	pu.postRepo.GetInfo()
}

func (pu *postUseCase) PostInfo() {
	fmt.Println("Post usecase PostInfo")
	pu.postRepo.PostInfo()
}
