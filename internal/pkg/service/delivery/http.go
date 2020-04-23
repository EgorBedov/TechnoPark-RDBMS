package delivery

import (
	"egogoger/internal/pkg/network"
	"egogoger/internal/pkg/service"
	"github.com/go-chi/chi"
	//"log"
	"net/http"
)

type ServiceHandler struct {
	serviceUseCase service.UseCase
}

func NewServiceHandler(fu service.UseCase, r *chi.Mux) {
	handler := &ServiceHandler{serviceUseCase:fu}

	r.Route("/api/service", func(r chi.Router) {
		r.Post("/clear", 	handler.TruncateAll)
		r.Get("/status", 	handler.GetInfo)
	})
}

func (sh *ServiceHandler) TruncateAll(w http.ResponseWriter, r *http.Request) {
	//log.Println("/service/status working")

	status := sh.serviceUseCase.TruncateAll()

	if status != http.StatusOK {
		network.GenErrorCode(w, r, http.StatusText(status), status)
		return
	}

	//log.Println("/service/clear worked nicely")
	network.Jsonify(w, http.StatusText(status), status)
}

func (sh *ServiceHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	//log.Println("/service/status working")

	summary, status := sh.serviceUseCase.GetInfo()

	if status != http.StatusOK {
		network.GenErrorCode(w, r, http.StatusText(status), status)
		return
	}

	//log.Println("/service/status worked nicely")
	network.Jsonify(w, &summary, status)
}
