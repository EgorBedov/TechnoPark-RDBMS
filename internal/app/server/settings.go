package server

import (
	"egogoger/internal/pkg/db"
	"egogoger/internal/pkg/forum/delivery"
	"egogoger/internal/pkg/forum/repository"
	"egogoger/internal/pkg/forum/usecase"
	"egogoger/internal/pkg/temp"
	"github.com/gorilla/mux"
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

var routesMap = map[string][]MapHandler {
	"/api/echo": {{
		Type: 		"GET",
		Handler:	temp.Echo,
	}},
}

func (ss *Settings) GetRouter() *mux.Router {
	conn := db.ConnectToDB()

	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	forum := api.PathPrefix("/forum").Subrouter()
	delivery.NewForumHandler(
		usecase.NewForumUseCase(
			repository.NewPgxForumRepository(conn)),
		forum)
	return r
}
