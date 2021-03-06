package server

import (
	"egogoger/internal/pkg/db"
	forumDelivery "egogoger/internal/pkg/forum/delivery"
	forumRepo "egogoger/internal/pkg/forum/repository"
	forumUseCase "egogoger/internal/pkg/forum/usecase"
	postDelivery "egogoger/internal/pkg/post/delivery"
	postRepo "egogoger/internal/pkg/post/repository"
	postUseCase "egogoger/internal/pkg/post/usecase"
	serviceDelivery "egogoger/internal/pkg/service/delivery"
	serviceRepo "egogoger/internal/pkg/service/repository"
	serviceUseCase "egogoger/internal/pkg/service/usecase"
	threadDelivery "egogoger/internal/pkg/thread/delivery"
	threadRepo "egogoger/internal/pkg/thread/repository"
	threadUseCase "egogoger/internal/pkg/thread/usecase"
	userDelivery "egogoger/internal/pkg/user/delivery"
	userRepository "egogoger/internal/pkg/user/repository"
	userUseCase "egogoger/internal/pkg/user/usecase"
	"github.com/go-chi/chi"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type MapHandler struct {
	Type         string
	Handler      HandlerFunc
}

type Settings struct {
	Port   int
	Ip     string
	Routes map[string][]MapHandler
	Router http.Handler
}

func GetConfig() Settings {
	return Settings{
		Port:   5000,
		Ip:     "0.0.0.0",
	}
}

func (ss *Settings) GetRouter() http.Handler {
	conn := db.ConnectToDB()

	r := chi.NewRouter()
	//r.Use(middleware.Logger)

	forumDelivery.NewForumHandler(
		forumUseCase.NewForumUseCase(
			forumRepo.NewPgxForumRepository(conn)),
		r)

	postDelivery.NewPostHandler(
		postUseCase.NewPostUseCase(
			postRepo.NewPgxPostRepository(conn)),
		r)

	serviceDelivery.NewServiceHandler(
		serviceUseCase.NewServiceUseCase(
			serviceRepo.NewPgxServiceRepository(conn)),
		r)

	threadDelivery.NewThreadHandler(
		threadUseCase.NewThreadUseCase(
			threadRepo.NewPgxThreadRepository(conn)),
		r)

	userDelivery.NewUserHandler(
		userUseCase.NewUserUseCase(
			userRepository.NewPgxUserRepository(conn)),
		r)
	return r
}
