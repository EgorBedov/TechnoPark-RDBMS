package network

import (
	"encoding/json"
	"net/http"
)

type Message struct {
	Request *http.Request `json:"-"`
	Message string        `json:"message"`
}

func Jsonify(w http.ResponseWriter, object interface{}, status int) {
	output, err := json.Marshal(object)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(output)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
}

func GenErrorCode(w http.ResponseWriter, r *http.Request, what string, status int) {
	w.WriteHeader(status)
	page := Message{r, what}
	output, err := json.Marshal(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write(output)
}
