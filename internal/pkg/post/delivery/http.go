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

	r.HandleFunc("/echo", handler.Echo).Methods("GET")
}

func (fh *PostHandler) Echo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Post Echo")

	fh.postUseCase.Echo()
}
