package v1

import (
	"net/http"
)

func SayHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}
