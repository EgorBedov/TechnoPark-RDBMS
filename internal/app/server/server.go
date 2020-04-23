package server

import (
	//"log"
	"net/http"
	"strconv"
)

func Start() {
	//file := logger.OpenLogFile()
	//defer file.Close()

	serverSettings := GetConfig()

	//log.Println("Server is running on " + serverSettings.Ip + ":" + strconv.Itoa(serverSettings.Port))
	if err := http.ListenAndServe(serverSettings.Ip + ":" + strconv.Itoa(serverSettings.Port), serverSettings.GetRouter()); err != nil {
		//log.Println("ERROR", err)
	}
}
