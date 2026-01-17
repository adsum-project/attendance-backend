package main

import (
	"net/http"

	"github.com/adsum-project/attendance-backend/pkg/router"
)

func main() {
	r := router.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World2!"))
	})

	r.StartServer(":8080")
}