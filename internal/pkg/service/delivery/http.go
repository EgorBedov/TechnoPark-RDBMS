package delivery

import (
	"egogoger/internal/pkg/service"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

type ServiceHandler struct {
	serviceUseCase service.UseCase
}

func NewServiceHandler(fu service.UseCase, r *mux.Router) {
	handler := &ServiceHandler{serviceUseCase:fu}

	r.HandleFunc("/echo", handler.Echo).Methods("GET")
}

func (fh *ServiceHandler) Echo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Service Echo")

	fh.serviceUseCase.Echo()
}
