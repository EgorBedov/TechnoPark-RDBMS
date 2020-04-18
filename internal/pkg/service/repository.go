package service

import "egogoger/internal/pkg/models"

type Repository interface {
	TruncateAll() int
	GetInfo() (*models.Summary, int)
}
