package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/user"
	"encoding/json"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type UserHandler struct {
	userUseCase user.UseCase
}

func NewUserHandler(fu user.UseCase, r *chi.Mux) {
	handler := &UserHandler{userUseCase:fu}

	r.Route("/api/user/{nickname}", func(r chi.Router) {
		r.Post("/create", 	handler.CreateUser)
		r.Get("/profile", 	handler.GetInfo)
		r.Post("/profile", 	handler.PostInfo)
	})
}

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("/user/{nickname}/create POST working")

	decoder := json.NewDecoder(r.Body)
	var usr models.User
	if err := decoder.Decode(&usr); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}
	usr.NickName = chi.URLParam(r, "nickname")

	users, status := uh.userUseCase.CreateUser(&usr)

	log.Println("/user/{nickname}/create POST worked nicely")
	if users[0] == usr {
		network.Jsonify(w, usr, status)
	} else {
		network.Jsonify(w, users, status)
	}
}

func (uh *UserHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("/user/{nickname}/profile GET working")
	usr := models.User{
		NickName: chi.URLParam(r, "nickname"),
	}
	status := uh.userUseCase.GetInfo(&usr)

	if status != http.StatusOK {
		network.GenErrorCode(w, r, "Can't find user by nickname: " + usr.NickName, status)
		return
	}

	log.Println("/user/{nickname}/profile GET worked nicely")
	network.Jsonify(w, usr, status)
}

func (uh *UserHandler) PostInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("/user/{nickname}/profile POST working")

	decoder := json.NewDecoder(r.Body)
	var usr models.User
	if err := decoder.Decode(&usr); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}
	usr.NickName = chi.URLParam(r, "nickname")

	status, message := uh.userUseCase.PostInfo(&usr)

	if status != http.StatusOK {
		network.GenErrorCode(w, r, message.Message, status)
		return
	}

	log.Println("/user/{nickname}/profile POST worked nicely")
	network.Jsonify(w, usr, status)
}
