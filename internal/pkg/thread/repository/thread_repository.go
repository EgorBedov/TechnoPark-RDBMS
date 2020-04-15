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

func (tr *threadRepository) CreatePosts() {
	fmt.Println("Thread repo CreatePosts")
}

func (tr *threadRepository) GetInfo() {
	fmt.Println("Thread repo GetInfo")
}

func (tr *threadRepository) PostInfo() {
	fmt.Println("Thread repo PostInfo")
}

func (tr *threadRepository) GetPosts() {
	fmt.Println("Thread repo GetPosts")
}

func (tr *threadRepository) Vote() {
	fmt.Println("Thread repo Vote")
}
