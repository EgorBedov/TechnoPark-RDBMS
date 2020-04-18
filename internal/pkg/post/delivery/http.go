package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/post"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type PostHandler struct {
	postUseCase post.UseCase
}

func NewPostHandler(fu post.UseCase, r *mux.Router) {
	handler := &PostHandler{postUseCase:fu}

	id := r.PathPrefix("/{id}").Subrouter()
	id.HandleFunc("/details", 	handler.GetInfo)	.Methods("GET")
	id.HandleFunc("/details", 	handler.PostInfo)	.Methods("POST")
}

func (ph *PostHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here")
	postInfoQuery, err := models.DecodePostInfoQuery(r)
	if err != nil {
		network.GenErrorCode(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	status, postInfo := ph.postUseCase.GetInfo(postInfoQuery)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find post with id " + strconv.Itoa(postInfoQuery.PostId), status)
		return
	}

	network.Jsonify(w, &postInfo, status)
}

func (ph *PostHandler) PostInfo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var msg models.Message
	if err := decoder.Decode(&msg); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}

	idString := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idString)
	if err != nil {
		network.GenErrorCode(w, r, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	pst, status := ph.postUseCase.PostInfo(id, msg)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find post with id " + idString, status)
		return
	}

	network.Jsonify(w, pst, status)
}
