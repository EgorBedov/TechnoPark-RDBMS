package forum

type Repository interface {
	CreateForum()
	GetInfo()
	CreateThread()
	GetUsers()
	GetThreads()
}
