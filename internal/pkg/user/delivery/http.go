package delivery

import (
	"egogoger/internal/pkg/user"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type UserHandler struct {
	userUseCase user.UseCase
}

func NewUserHandler(fu user.UseCase, r *mux.Router) {
	handler := &UserHandler{userUseCase:fu}

	r.HandleFunc("/echo", handler.Echo).Methods("GET")
}

func (fh *UserHandler) Echo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("User Echo")

	fh.userUseCase.Echo()
}
