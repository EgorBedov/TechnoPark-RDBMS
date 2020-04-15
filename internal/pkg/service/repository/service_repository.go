package repository

import (
	"egogoger/internal/pkg/service"
	"fmt"
	"github.com/jackc/pgx"
)

type serviceRepository struct {
	db *pgx.ConnPool
}

func NewPgxServiceRepository(db *pgx.ConnPool) service.Repository {
	return &serviceRepository{db: db}
}

func (fr *serviceRepository) Echo() {
	fmt.Println("Service repo")
}
