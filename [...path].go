package handler

import (
	"net/http"

	"tukychat/pkg/web"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	web.Handler().ServeHTTP(w, r)
}
