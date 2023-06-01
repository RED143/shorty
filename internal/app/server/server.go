package server

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"shorty/internal/app/compress"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
	"shorty/internal/app/logger"
	"shorty/internal/app/storage"
)

type handler struct {
	config  config.Config
	storage storage.Storage
	logger  logger.Logger
}

func (h *handler) getLink(writer http.ResponseWriter, request *http.Request) {
	hash := chi.URLParam(request, "hash")
	handlers.GetLink(writer, request, hash, h.storage, h.logger)
}

func (h *handler) shortifyLink(writer http.ResponseWriter, request *http.Request) {
	handlers.Shortify(writer, request, h.config, h.storage, h.logger)
}

func (h *handler) shortenLink(writer http.ResponseWriter, request *http.Request) {
	handlers.ShortenLink(writer, request, h.config, h.storage, h.logger)
}

type middleware struct {
	logger logger.Logger
}

func (m *middleware) withLogging(h http.Handler) http.Handler {
	return logger.WithLogging(h, m.logger)
}

func (m *middleware) WithCompressing(h http.Handler) http.Handler {
	return compress.WithCompressing(h, m.logger)
}

func Start() error {
	logger, err := logger.Initialize()
	c := config.GetConfig()
	h := handler{storage: storage.NewStorage(c.FileStoragePath), config: c, logger: logger}
	m := middleware{logger: logger}
	router := chi.NewRouter()
	if err != nil {
		return err
	}

	router.Use(m.withLogging)
	router.Use(m.WithCompressing)

	router.Post("/", h.shortifyLink)
	router.Post("/api/shorten", h.shortenLink)
	router.Get("/{hash}", h.getLink)

	logger.Infow("Starting server", "address", h.config.ServerAddress)
	if err := http.ListenAndServe(h.config.ServerAddress, router); err != nil {
		return err
	}
	return nil
}
