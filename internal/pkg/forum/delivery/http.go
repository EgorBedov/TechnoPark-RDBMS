package delivery

import (
	"egogoger/internal/pkg/forum"
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"encoding/json"
	"github.com/go-chi/chi"
	//"log"
	"net/http"
)

type ForumHandler struct {
	forumUseCase forum.UseCase
}

func NewForumHandler(fu forum.UseCase, r *chi.Mux) {
	handler := &ForumHandler{forumUseCase:fu}

	r.Route("/api/forum", func(r chi.Router) {
		r.Post("/create", handler.CreateForum)
		r.Route("/{slug}", func(r chi.Router) {
			r.Get("/details", 	handler.GetInfo)
			r.Post("/create", 	handler.CreateThread)
			r.Get("/users", 	handler.GetUsers)
			r.Get("/threads", 	handler.GetThreads)
		})
	})
}

func (fh *ForumHandler) CreateForum(w http.ResponseWriter, r *http.Request) {
	//log.Println("/forum/create working ")

	decoder := json.NewDecoder(r.Body)
	var foroom models.Forum
	if err := decoder.Decode(&foroom); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}

	status, message := fh.forumUseCase.CreateForum(&foroom)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, message.Message, status)
		return
	}

	//log.Println("/forum/create worked nicely ")
	network.Jsonify(w, foroom, status)
}

func (fh *ForumHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	//log.Println("/forum/{slug}/details GET working ")

	frm := models.Forum{
		Slug: chi.URLParam(r, "slug"),
	}
	status := fh.forumUseCase.GetInfo(&frm)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find forum with slug " + frm.Slug, status)
		return
	}

	//log.Println("/forum/{slug}/details GET worked nicely ")
	network.Jsonify(w, frm, status)
}

func (fh *ForumHandler) CreateThread(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var thrd models.Thread
	if err := decoder.Decode(&thrd); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}
	thrd.Forum = chi.URLParam(r, "slug")
	thrd.Created = thrd.Created.UTC()

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

	if users == nil {
		users = []models.User{}
	}
	network.Jsonify(w, users, status)
}

func (fh *ForumHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	//log.Println("/forum/{slug}/threads GET working")

	query := models.DecodeQuery(r)
	threads, status := fh.forumUseCase.GetThreads(query)
	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find forum with slug " + query.Slug, status)
		return
	}

	//log.Println("/forum/{slug}/threads GET worked nicely ")

	if threads == nil {
		threads = []models.Thread{}
	}
	network.Jsonify(w, threads, status)
}
