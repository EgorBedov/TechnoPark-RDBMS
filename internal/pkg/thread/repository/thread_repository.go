package repository

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/thread"
	"fmt"
	"github.com/jackc/pgx"
	"net/http"
)

type threadRepository struct {
	db *pgx.ConnPool
}

func NewPgxThreadRepository(db *pgx.ConnPool) thread.Repository {
	return &threadRepository{db: db}
}

func (tr *threadRepository) CreatePosts(posts []models.Post, slugOrId string) int {
	//sqlStatement := `
	//	INSERT INTO post (id, parent, author, message, isEdited, forum, thread_id, created) VALUES
	//		();`

	return http.StatusOK
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
