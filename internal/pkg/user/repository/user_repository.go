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

func (fr *userRepository) Echo() {
	fmt.Println("User repo")
}
