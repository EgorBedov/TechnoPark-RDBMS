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

	r.HandleFunc("/echo", handler.Echo).Methods("GET")
}

func (fh *ForumHandler) Echo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Forum Echo")

	fh.forumUseCase.Echo()
}
