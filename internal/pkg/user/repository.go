package user

type Repository interface {
	CreateUser()
	GetInfo()
	PostInfo()
}
