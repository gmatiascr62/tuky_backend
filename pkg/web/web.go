package web

import (
	"net/http"

	"tukychat/internal/app"
)

func Handler() http.Handler {
	return app.NewRouter()
}
