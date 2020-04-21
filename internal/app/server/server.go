package server

import (
	"egogoger/internal/pkg/logger"
	"log"
	"net/http"
	"strconv"
)

func Start() {
	file := logger.OpenLogFile()
	defer file.Close()

	serverSettings := GetConfig()
	//server := http.Server{
	//	Addr:              serverSettings.Ip + ":" + strconv.Itoa(serverSettings.Port),
	//	Handler: 		   serverSettings.GetRouter(),
	//	ReadTimeout:       0,
	//	ReadHeaderTimeout: 0,
	//	WriteTimeout:      0,
	//	IdleTimeout:       0,
	//	MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	//}

	log.Println("Server is running on " + strconv.Itoa(serverSettings.Port))
	if err := http.ListenAndServe(serverSettings.Ip + ":" + strconv.Itoa(serverSettings.Port), serverSettings.GetRouter()); err != nil {
		log.Println(err)
	}
}
