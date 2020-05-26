package delivery

import (
	"egogoger/internal/pkg/models"
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/post"
	"encoding/json"
	"github.com/go-chi/chi"
	//"log"
	"net/http"
	"strconv"
)

type PostHandler struct {
	postUseCase post.UseCase
}

func NewPostHandler(fu post.UseCase, r *chi.Mux) {
	handler := &PostHandler{postUseCase:fu}

	r.Route("/api/post/{id}", func(r chi.Router) {
		r.Get("/details", 	handler.GetInfo)
		r.Post("/details", 	handler.PostInfo)
	})
}

func (ph *PostHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
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
	//log.Println("/post/{id}/details POST working")

	decoder := json.NewDecoder(r.Body)
	var pst models.Post
	if err := decoder.Decode(&pst); err != nil {
		network.GenErrorCode(w, r, "Error within parse json", http.StatusBadRequest)
		return
	}

	idString := chi.URLParam(r, "id")
	var err error
	pst.Id, err = strconv.Atoi(idString)
	if err != nil {
		network.GenErrorCode(w, r, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	status := ph.postUseCase.PostInfo(&pst)

	if status == http.StatusNotFound {
		network.GenErrorCode(w, r, "Can't find post with id " + idString, status)
		return
	}

	//log.Println("/post/{id}/details POST worked nicely")
	network.Jsonify(w, pst, status)
}
