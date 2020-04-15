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

	r.HandleFunc("/clear", 	handler.TruncateAll)	.Methods("POST")
	r.HandleFunc("/status", 	handler.GetInfo)		.Methods("GET")
}

func (sh *ServiceHandler) TruncateAll(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Service handler TruncateAll")
	sh.serviceUseCase.TruncateAll()
}

func (sh *ServiceHandler) GetInfo(h http.ResponseWriter, r *http.Request) {
	fmt.Println("Service handler GetInfo")
	sh.serviceUseCase.GetInfo()
}
