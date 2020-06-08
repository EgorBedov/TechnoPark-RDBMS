package server

import (
	"egogoger/internal/pkg/cache"
	//"log"
	"net/http"
	"strconv"
)

func Start() {
	serverSettings := GetConfig()

	router := serverSettings.GetRouter()

	cache.FillForums()
	cache.InitThreadsCaches()

	_ = http.ListenAndServe(serverSettings.Ip + ":" + strconv.Itoa(serverSettings.Port), router)
}
