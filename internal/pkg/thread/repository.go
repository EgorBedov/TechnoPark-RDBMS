package thread

type Repository interface {
	CreatePosts()
	GetInfo()
	PostInfo()
	GetPosts()
	Vote()
}
