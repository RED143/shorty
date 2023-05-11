package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
)

func Start() error {
	config.ParseFlags()
	cfg := config.GetConfig()
	router := chi.NewRouter()

	router.Post("/", handlers.Shortify)
	router.Get("/{hash}", handlers.GetLink)

	err := http.ListenAndServe(cfg.ServerAddress, router)

	return err
}
