package delivery

import (
	"egogoger/internal/pkg/thread"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ThreadHandler struct {
	threadUseCase thread.UseCase
}

func NewThreadHandler(fu thread.UseCase, r *mux.Router) {
	handler := &ThreadHandler{threadUseCase:fu}

	r.HandleFunc("/echo", handler.Echo).Methods("GET")
}

func (fh *ThreadHandler) Echo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Thread Echo")

	fh.threadUseCase.Echo()
}
