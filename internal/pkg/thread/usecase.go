package thread

type UseCase interface {
	CreatePosts()
	GetInfo()
	PostInfo()
	GetPosts()
	Vote()
}