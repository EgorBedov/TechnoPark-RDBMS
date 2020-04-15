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

	nickname := r.PathPrefix("/{nickname}").Subrouter()
	nickname.HandleFunc("/create", 	handler.CreateUser)	.Methods("POST")
	nickname.HandleFunc("/profile", 	handler.GetInfo)	.Methods("GET")
	nickname.HandleFunc("/profile", 	handler.PostInfo)	.Methods("POST")
}

func (uh *UserHandler) CreateUser(h http.ResponseWriter, r *http.Request) {
	fmt.Println("User CreateUser")
	uh.userUseCase.CreateUser()
}

func (uh *UserHandler) GetInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("User GetInfo")
	uh.userUseCase.GetInfo()
}

func (uh *UserHandler) PostInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("User PostInfo")
	uh.userUseCase.PostInfo()
}
