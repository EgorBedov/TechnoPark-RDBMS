package forum

type UseCase interface {
	CreateForum()
	GetInfo()
	CreateThread()
	GetUsers()
	GetThreads()
}