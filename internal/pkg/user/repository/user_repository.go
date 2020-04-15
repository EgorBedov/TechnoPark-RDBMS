package repository

import (
	"egogoger/internal/pkg/user"
	"fmt"
	"github.com/jackc/pgx"
)

type userRepository struct {
	db *pgx.ConnPool
}

func NewPgxUserRepository(db *pgx.ConnPool) user.Repository {
	return &userRepository{db: db}
}

func (ur *userRepository) CreateUser() {
	fmt.Println("User repo CreateUser")
}

func (ur *userRepository) GetInfo() {
	fmt.Println("User repo GetInfo")
}

func (ur *userRepository) PostInfo() {
	fmt.Println("User repo PostInfo")
}
