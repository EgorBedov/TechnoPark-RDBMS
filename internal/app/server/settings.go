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
	"github.com/gorilla/mux"
	"log"
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
		Ip:     "127.0.0.1",
	}
}

func (ss *Settings) GetRouter() *mux.Router {
	conn := db.ConnectToDB()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	forumDelivery.NewForumHandler(
		forumUseCase.NewForumUseCase(
			forumRepo.NewPgxForumRepository(conn)),
		api.PathPrefix("/forum").Subrouter())

	postDelivery.NewPostHandler(
		postUseCase.NewPostUseCase(
			postRepo.NewPgxPostRepository(conn)),
		api.PathPrefix("/post").Subrouter())

	serviceDelivery.NewServiceHandler(
		serviceUseCase.NewServiceUseCase(
			serviceRepo.NewPgxServiceRepository(conn)),
		api.PathPrefix("/service").Subrouter())

	threadDelivery.NewThreadHandler(
		threadUseCase.NewThreadUseCase(
			threadRepo.NewPgxThreadRepository(conn)),
		api.PathPrefix("/thread").Subrouter())

	userDelivery.NewUserHandler(
		userUseCase.NewUserUseCase(
			userRepository.NewPgxUserRepository(conn)),
		api.PathPrefix("/user").Subrouter())
	log.Println("Handlers were set")
	return r
}
