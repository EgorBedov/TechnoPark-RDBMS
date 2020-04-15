package repository

import (
	"egogoger/internal/pkg/thread"
	"fmt"
	"github.com/jackc/pgx"
)

type threadRepository struct {
	db *pgx.ConnPool
}

func NewPgxThreadRepository(db *pgx.ConnPool) thread.Repository {
	return &threadRepository{db: db}
}

func (fr *threadRepository) Echo() {
	fmt.Println("Thread repo")
}
