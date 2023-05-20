package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
	"shorty/internal/app/logger"
	"shorty/internal/app/storage"
)

type handler struct {
	config  config.Config
	storage *storage.Storage
}

func Start() error {
	h := handler{storage: storage.NewStorage(), config: config.GetConfig()}
	router := chi.NewRouter()
	logger.Initialize()

	router.Post("/", logger.WithLogging(h.shortifyHandler))
	router.Get("/{hash}", logger.WithLogging(h.getLinkHandler))

	logger.Info("Starting server", "address", h.config.ServerAddress)
	if err := http.ListenAndServe(h.config.ServerAddress, router); err != nil {
		return err
	}
	return nil
}

func (h *handler) getLinkHandler(writer http.ResponseWriter, request *http.Request) {
	hash := chi.URLParam(request, "hash")
	handlers.GetLink(writer, request, hash, h.storage)
}

func (h *handler) shortifyHandler(writer http.ResponseWriter, request *http.Request) {
	handlers.Shortify(writer, request, h.config, h.storage)
}
