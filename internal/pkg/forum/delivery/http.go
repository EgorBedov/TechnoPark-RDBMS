package delivery

import (
	"egogoger/internal/pkg/forum"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ForumHandler struct {
	forumUseCase forum.UseCase
}

func NewForumHandler(fu forum.UseCase, r *mux.Router) {
	handler := &ForumHandler{forumUseCase:fu}

	r.HandleFunc("/create", 		handler.CreateForum)	.Methods("POST")
	slug := r.PathPrefix("/{slug}").Subrouter()
	slug.HandleFunc("/details", 	handler.GetInfo)		.Methods("GET")
	slug.HandleFunc("/create", 	handler.CreateThread)	.Methods("POST")
	slug.HandleFunc("/users", 	handler.GetUsers)		.Methods("GET")
	slug.HandleFunc("/threads", 	handler.GetThreads)		.Methods("GET")
}

func (fh *ForumHandler) CreateForum(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Forum handler CreateForum()")
	fh.forumUseCase.CreateForum()
}

func (fh *ForumHandler) GetInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Forum handler GetInfo()")
	fh.forumUseCase.GetInfo()
}

func (fh *ForumHandler) CreateThread(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Forum handler CreateThread()")
	fh.forumUseCase.CreateThread()
}

func (fh *ForumHandler) GetUsers(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Forum handler GetUsers()")
	fh.forumUseCase.GetUsers()
}

func (fh *ForumHandler) GetThreads(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Forum handler GetThreads()")
	fh.forumUseCase.GetThreads()
}
