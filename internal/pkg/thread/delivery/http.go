package delivery

import (
	"egogoger/internal/pkg/thread"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ThreadHandler struct {
	threadUseCase thread.UseCase
}

func NewThreadHandler(fu thread.UseCase, r *mux.Router) {
	handler := &ThreadHandler{threadUseCase:fu}

	slugOrId := r.PathPrefix("/{slug_or_id}").Subrouter()
	slugOrId.HandleFunc("/create", 	handler.CreatePosts)	.Methods("POST")
	slugOrId.HandleFunc("/details", 	handler.GetInfo)		.Methods("GET")
	slugOrId.HandleFunc("/details", 	handler.PostInfo)		.Methods("POST")
	slugOrId.HandleFunc("/posts", 	handler.GetPosts)		.Methods("GET")
	slugOrId.HandleFunc("/vote", 		handler.Vote)			.Methods("POST")
}

func (th *ThreadHandler) CreatePosts(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler CreatePosts")
	th.threadUseCase.CreatePosts()
}

func (th *ThreadHandler) GetInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler GetInfo")
	th.threadUseCase.GetInfo()
}

func (th *ThreadHandler) PostInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler PostInfo")
	th.threadUseCase.PostInfo()
}

func (th *ThreadHandler) GetPosts(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler GetPosts")
	th.threadUseCase.GetPosts()
}

func (th *ThreadHandler) Vote(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler Vote")
	th.threadUseCase.Vote()
}
