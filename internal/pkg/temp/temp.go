package temp

import (
	"fmt"
	"net/http"
)

func Echo(h http.ResponseWriter, r *http.Request) {

	fmt.Println("lol")
	//network.Jsonify(w, events, http.StatusOK)
}
