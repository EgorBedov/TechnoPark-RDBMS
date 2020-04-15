package repository

import (
	"egogoger/internal/pkg/post"
	"fmt"
	"github.com/jackc/pgx"
)

type postRepository struct {
	db *pgx.ConnPool
}

func NewPgxPostRepository(db *pgx.ConnPool) post.Repository {
	return &postRepository{db: db}
}

func (fr *postRepository) Echo() {
	fmt.Println("Post repo")
}
