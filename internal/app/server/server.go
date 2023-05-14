package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
)

func Start() error {
	cfg := config.GetConfig()
	router := chi.NewRouter()

	router.Post("/", shortifyHandler(cfg))
	router.Get("/{hash}", getLinkHandler)

	err := http.ListenAndServe(cfg.ServerAddress, router)

	return err
}

func getLinkHandler(writer http.ResponseWriter, request *http.Request) {
	hash := chi.URLParam(request, "hash")
	handlers.GetLink(writer, request, hash)
}

func shortifyHandler(config config.Config) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handlers.Shortify(writer, request, config)
	}
}
