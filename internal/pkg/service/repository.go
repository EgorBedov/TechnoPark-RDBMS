package service

type Repository interface {
	TruncateAll()
	GetInfo()
}
