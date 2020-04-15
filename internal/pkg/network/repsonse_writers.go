package network

import (
	"encoding/json"
	"log"
	"net/http"
)

func Jsonify(w http.ResponseWriter, object interface{}, status int) {
	output, err := json.Marshal(object)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(output)
	if err != nil {
		http.Error(w, err.Error(), status)
		return
	}
	log.Println("Sent json")
}
