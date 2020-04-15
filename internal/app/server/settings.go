package server

import (
	"forum/internal/pkg/temp"
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
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/echo", temp.Echo).
		Methods("GET")
	return r
}
