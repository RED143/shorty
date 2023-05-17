package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
	"shorty/internal/app/storage"
)

func Start() error {
	cfg := config.GetConfig()
	router := chi.NewRouter()
	str := storage.NewStorage()

	router.Post("/", shortifyHandler(cfg, str))
	router.Get("/{hash}", getLinkHandler(str))

	if err := http.ListenAndServe(cfg.ServerAddress, router); err != nil {
		return err
	}
	return nil
}

func getLinkHandler(str storage.Storage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		hash := chi.URLParam(request, "hash")
		handlers.GetLink(writer, request, hash, str)
	}
}

func shortifyHandler(config config.Config, str storage.Storage) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handlers.Shortify(writer, request, config, str)
	}
}
