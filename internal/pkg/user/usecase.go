package user

type UseCase interface {
	CreateUser()
	GetInfo()
	PostInfo()
}