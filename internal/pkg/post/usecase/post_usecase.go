package usecase

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/post"
)

type postUseCase struct {
	postRepo		post.Repository
}

func NewPostUseCase(f post.Repository) post.UseCase {
	return &postUseCase{postRepo: f}
}

func (pu *postUseCase) GetInfo(query *models.PostInfoQuery) (int, *models.PostInfo) {
	return pu.postRepo.GetInfo(query)
}

func (pu *postUseCase) PostInfo(postId int, msg models.Message) (*models.Post, int) {
	return pu.postRepo.PostInfo(postId, msg)
}
