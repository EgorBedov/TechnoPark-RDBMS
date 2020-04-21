package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/thread"
	"encoding/json"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type ThreadHandler struct {
	threadUseCase thread.UseCase
}

func NewThreadHandler(fu thread.UseCase, r *chi.Mux) {
	handler := &ThreadHandler{threadUseCase:fu}

	r.Route("/api/thread/{slug_or_id}", func(r chi.Router) {
		r.Post("/create", 	handler.CreatePosts)
		r.Get("/details", 	handler.GetInfo)
		r.Post("/details", 	handler.UpdateThread)
		r.Get("/posts", 	handler.GetPosts)
		r.Post("/vote", 	handler.Vote)
	})
}

func (th *ThreadHandler) CreatePosts(w http.ResponseWriter, r *http.Request) {
	log.Println("/thread/{slug_or_id}/create POST working")
	decoder := json.NewDecoder(r.Body)
	var posts []models.Post
	if err := decoder.Decode(&posts); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}

	threadId, err := th.threadUseCase.GetThreadIdBySlugOrId(chi.URLParam(r, "slug_or_id"))
	if err != nil {
		network.Jsonify(w, models.Post{}, http.StatusNotFound)
		return
	}

	if len(posts) == 0 {
		log.Println("/thread/{slug_or_id}/create POST finished with empty ")
		network.Jsonify(w, posts, http.StatusCreated)
		return
	}

	slugOrId := chi.URLParam(r, "slug_or_id")
	status := th.threadUseCase.CreatePosts(posts, threadId)

	if status != http.StatusCreated {
		log.Println("/thread/{slug_or_id}/create POST finished with error ", status)
		network.GenErrorCode(w, r, "Can't find parent message or thread with slug_or_id " + slugOrId, status)
		return
	}

	log.Println("/thread/{slug_or_id}/create POST worked nicely ")
	network.Jsonify(w, posts, status)
}

func (th *ThreadHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("/thread/{slug_or_id}/details GET working")

	slugOrId := chi.URLParam(r, "slug_or_id")
	thrd := models.Thread{}
	status := th.threadUseCase.GetInfo(&thrd, slugOrId)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug or id " + slugOrId, status)
		return
	}

	log.Println("/thread/{slug_or_id}/details GET worked nicely ")
	network.Jsonify(w, thrd, status)
}

func (th *ThreadHandler) UpdateThread(w http.ResponseWriter, r *http.Request) {
	log.Println("/thread/{slug_or_id}/details POST working")

	decoder := json.NewDecoder(r.Body)
	var thrd models.Thread
	if err := decoder.Decode(&thrd); err != nil {
		network.GenErrorCode(w, r, err.Error(), http.StatusBadRequest)
		return
	}
	slugOrId := chi.URLParam(r, "slug_or_id")

	status := th.threadUseCase.UpdateThread(&thrd, slugOrId)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug_or_id " + slugOrId, status)
		return
	}

	log.Println("/thread/{slug_or_id}/details POST worked nicely ")
	network.Jsonify(w, thrd, status)
}

func (th *ThreadHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	log.Println("/thread/{slug_or_id}/create GET working")

	query := models.DecodePostQuery(r)
	posts, status := th.threadUseCase.GetPosts(&query)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find thread with slug or id " + query.SlugOrId, status)
		return
	}

	log.Println("/thread/{slug_or_id}/create GET worked nicely")

	if posts == nil {
		posts = []models.Post{}
	}
	network.Jsonify(w, posts, status)
}

func (th *ThreadHandler) Vote(w http.ResponseWriter, r *http.Request) {
	log.Println("/thread/{slug_or_id}/vote POST working")

	// Get info about thread
	slugOrId := chi.URLParam(r, "slug_or_id")
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
	status, thrd.Votes = th.threadUseCase.Vote(&vote)
	if status != http.StatusOK {
		network.GenErrorCode(w, r, "Vote failed", status)
		return
	}

	log.Println("/thread/{slug_or_id}/vote POST worked nicely ")
	network.Jsonify(w, thrd, status)
}


