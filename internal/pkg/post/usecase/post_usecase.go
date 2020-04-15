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

func (fu *postUseCase) Echo() {
	fmt.Println("Post usecase")

	fu.postRepo.Echo()
}
