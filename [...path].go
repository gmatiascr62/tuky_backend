package handler

import (
	"net/http"

	"tukychat/pkg/web"
)

var handler = web.Handler()

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}
