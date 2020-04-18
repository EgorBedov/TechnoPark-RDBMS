package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/thread"
	"encoding/json"
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
	slugOrId.HandleFunc("/details", 	handler.UpdateThread)	.Methods("POST")
	slugOrId.HandleFunc("/posts", 	handler.GetPosts)		.Methods("GET")
	slugOrId.HandleFunc("/vote", 		handler.Vote)			.Methods("POST")
}

func (th *ThreadHandler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var posts []models.Post
	if err := decoder.Decode(&posts); err != nil {
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
	slugOrId := mux.Vars(r)["slug_or_id"]
	thrd := models.Thread{}
	status := th.threadUseCase.GetInfo(&thrd, slugOrId)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug or id " + slugOrId, status)
		return
	}

	network.Jsonify(w, thrd, status)
}

func (th *ThreadHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var thrd models.Thread
	if err := decoder.Decode(&thrd); err != nil {
		network.GenErrorCode(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	slugOrId := mux.Vars(r)["slug_or_id"]

	status := th.threadUseCase.UpdateThread(&thrd, slugOrId)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug_or_id " + thrd.Slug, status)
		return
	}

	network.Jsonify(w, thrd, status)
}

func (th *ThreadHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	query := models.DecodePostQuery(r)
	posts, status := th.threadUseCase.GetPosts(&query)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug or id " + query.SlugOrId, status)
		return
	}

	network.Jsonify(w, posts, status)
}

func (th *ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {
	// Get info about thread
	slugOrId := mux.Vars(r)["slug_or_id"]
	thrd := models.Thread{}
	status := th.threadUseCase.GetInfo(&thrd, slugOrId)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug or id " + slugOrId, status)
		return
	}

	// Prepare vote
	decoder := json.NewDecoder(r.Body)
	var vote models.Vote
	if err := decoder.Decode(&vote); err != nil {
		network.GenErrorCode(w, r, err.Error(), http.StatusBadRequest)
		return
	}

	// Send vote
	vote.ThreadId = thrd.Id
	status = th.threadUseCase.Vote(&vote)
	if status != http.StatusOK {
		network.GenErrorCode(w, r, "Vote failed", status)
		return
	}

	// Change thrd and return
	thrd.Votes += vote.Voice
	network.Jsonify(w, thrd, status)
}
