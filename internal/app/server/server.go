package server

import (
	"net/http"

	"github.com/RED143/shorty/internal/app/config"
	"github.com/RED143/shorty/internal/app/handlers"
	"github.com/go-chi/chi/v5"
)

func Start() {
	config.InitConfig()
	router := chi.NewRouter()

	router.Post("/", handlers.ShortifyHandler)
	router.Get("/{hash}", handlers.LinkHandler)

	err := http.ListenAndServe(config.GetServerAddress(), router)

	if err != nil {
		panic(err)
	}
}
