package service

import "egogoger/internal/pkg/models"

type UseCase interface {
	TruncateAll() int
	GetInfo() (*models.Summary, int)
}