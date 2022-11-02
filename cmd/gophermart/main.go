package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	router := chi.NewRouter()
	router.Use(middleware.Compress(5))

	// router.Use(AuthMw)
	// router.Use(middleware.Compress(5))
	log.Panic(http.ListenAndServe("/", router))
}
