package delivery

import (
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ForumHandler struct {
	forumUseCase forum.UseCase
}

func NewForumHandler(fu forum.UseCase, r *mux.Router) {
	handler := &ForumHandler{forumUseCase:fu}

	r.HandleFunc("/create", 		handler.CreateForum)	.Methods("POST")
	slug := r.PathPrefix("/{slug}").Subrouter()
	slug.HandleFunc("/details", 	handler.GetInfo)		.Methods("GET")
	slug.HandleFunc("/create", 	handler.CreateThread)	.Methods("POST")
	slug.HandleFunc("/users", 	handler.GetUsers)		.Methods("GET")
	slug.HandleFunc("/threads", 	handler.GetThreads)		.Methods("GET")
}

func (fh *ForumHandler) CreateForum(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var foroom models.Forum
	if err := decoder.Decode(&foroom); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}

	status := fh.forumUseCase.CreateForum(&foroom)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find user with nickname " + foroom.Usr, status)
		return
	}
	network.Jsonify(w, foroom, status)
}

func (fh *ForumHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	frm := models.Forum{
		Slug: mux.Vars(r)["slug"],
	}
	status := fh.forumUseCase.GetInfo(&frm)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find forum with slug " + frm.Slug, status)
		return
	}

	network.Jsonify(w, frm, status)
}

func (fh *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var thrd models.Thread
	if err := decoder.Decode(&thrd); err != nil {
		fmt.Println(err)
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}
	thrd.Forum = mux.Vars(r)["slug"]

	status := fh.forumUseCase.CreateThread(&thrd)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find forum with slug " + thrd.Forum, status)
		return
	}

	network.Jsonify(w, thrd, status)
}

func (fh *ForumHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	query := models.DecodeQuery(r)
	users, status := fh.forumUseCase.GetUsers(query)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find forum with slug " + query.Slug, status)
		return
	}

	network.Jsonify(w, users, status)
}

func (fh *ForumHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	query := models.DecodeQuery(r)
	fmt.Println(query)
	threads, status := fh.forumUseCase.GetThreads(query)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find forum with slug " + query.Slug, status)
		return
	}

	network.Jsonify(w, threads, status)
}
