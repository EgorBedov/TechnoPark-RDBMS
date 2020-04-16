package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/thread"
	"encoding/json"
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

func (th *ThreadHandler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var posts []models.Post
	if err := decoder.Decode(&posts); err != nil {
		fmt.Println(err)
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}

	slugOrId := mux.Vars(r)["slug_or_id"]
	status := th.threadUseCase.CreatePosts(posts, slugOrId)

	if status != http.StatusOK {
		network.GenErrorCode(w, r, "Can't find parent message or thread with slug_or_id " + slugOrId, status)
		return
	}

	network.Jsonify(w, posts, status)
}

func (th *ThreadHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler GetInfo")
	th.threadUseCase.GetInfo()
}

func (th *ThreadHandler) PostInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler PostInfo")
	th.threadUseCase.PostInfo()
}

func (th *ThreadHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler GetPosts")
	th.threadUseCase.GetPosts()
}

func (th *ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread handler Vote")
	th.threadUseCase.Vote()
}
