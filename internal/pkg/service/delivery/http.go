package delivery

import (
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/service"
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

func (sh *ServiceHandler) TruncateAll(w http.ResponseWriter, r *http.Request) {
	status := sh.serviceUseCase.TruncateAll()

	if status != http.StatusOK {
		network.GenErrorCode(w, r, http.StatusText(status), status)
		return
	}

	network.Jsonify(w, http.StatusText(status), status)
}

func (sh *ServiceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	summary, status := sh.serviceUseCase.GetInfo()

	if status != http.StatusOK {
		network.GenErrorCode(w, r, http.StatusText(status), status)
		return
	}

	network.Jsonify(w, &summary, status)
}
