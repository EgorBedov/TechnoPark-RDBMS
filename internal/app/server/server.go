package server

import (
	"fmt"
	"net/http"
	"strconv"
)

func Start() {
	serverSettings := GetConfig()
	server := http.Server{
		Addr:              serverSettings.Ip + ":" + strconv.Itoa(serverSettings.Port),
		Handler: 		   serverSettings.GetRouter(),
		ReadTimeout:       0,
		ReadHeaderTimeout: 0,
		WriteTimeout:      0,
		IdleTimeout:       0,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
