package main

import (
	"tukychat/internal/app"
)

func main() {
	r := app.NewRouter()
	r.Run(":8080")
}