package delivery

import (
	"egogoger/internal/pkg/post"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type PostHandler struct {
	postUseCase post.UseCase
}

func NewPostHandler(fu post.UseCase, r *mux.Router) {
	handler := &PostHandler{postUseCase:fu}

	r.HandleFunc("/{id}/details", 	handler.GetInfo)	.Methods("GET")
	r.HandleFunc("/{id}/details", 	handler.PostInfo)	.Methods("POST")
}

func (ph *PostHandler) GetInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Post handler GetInfo")
	ph.postUseCase.GetInfo()
}

func (ph *PostHandler) PostInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Post handler PostInfo")
	ph.postUseCase.PostInfo()
}
