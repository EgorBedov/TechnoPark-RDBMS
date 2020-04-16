package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/user"
	"encoding/json"
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

func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var user models.User
	if err := decoder.Decode(&user); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}
	user.NickName = mux.Vars(r)["nickname"]

	status := uh.userUseCase.CreateUser(&user)
	network.Jsonify(w, user, status)
}

func (uh *UserHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	usr := models.User{
		NickName: mux.Vars(r)["nickname"],
	}
	status := uh.userUseCase.GetInfo(&usr)

	if status != http.StatusOK {
		network.GenErrorCode(w, r, "Can't find user with nickname " + usr.NickName, status)
		return
	}

	network.Jsonify(w, usr, status)
}

func (uh *UserHandler) PostInfo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var usr models.User
	if err := decoder.Decode(&usr); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}
	usr.NickName = mux.Vars(r)["nickname"]

	status := uh.userUseCase.PostInfo(&usr)

	if status != http.StatusOK {
		network.GenErrorCode(w, r, "Can't find user with nickname " + usr.NickName, status)
		return
	}
	network.Jsonify(w, usr, status)
}
