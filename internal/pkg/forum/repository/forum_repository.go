package repository

import (
	"egogoger/internal/pkg/forum"
	"fmt"
	"github.com/jackc/pgx"
)

type forumRepository struct {
	db *pgx.ConnPool
}

func NewPgxForumRepository(db *pgx.ConnPool) forum.Repository {
	return &forumRepository{db: db}
}

func (fr *forumRepository) Echo() {
	fmt.Println("Forum repo")
}
