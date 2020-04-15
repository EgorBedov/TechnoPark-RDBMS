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

func (fr *forumRepository) CreateForum() {
	fmt.Println("Forum repo CreateForum")
}

func (fr *forumRepository) GetInfo() {
	fmt.Println("Forum repo GetInfo")
}

func (fr *forumRepository) CreateThread() {
	fmt.Println("Forum repo CreateThread")
}

func (fr *forumRepository) GetUsers() {
	fmt.Println("Forum repo GetUsers")
}

func (fr *forumRepository) GetThreads() {
	fmt.Println("Forum repo GetThreads")
}
