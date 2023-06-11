package server

import (
	"net/http"
	"shorty/internal/app/compress"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
	"shorty/internal/app/logger"
	"shorty/internal/app/storage"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type handler struct {
	config  config.Config
	storage storage.Storage
	logger  *zap.SugaredLogger
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

func (h *handler) shortenLinkBatch(writer http.ResponseWriter, request *http.Request) {
	handlers.ShortenLinkBatch(writer, request, h.config, h.storage, h.logger)
}

func (h *handler) checkDatabaseConnection(writer http.ResponseWriter, request *http.Request) {
	handlers.CheckDatabaseConnection(writer, request, h.storage, h.logger)
}

type middleware struct {
	logger *zap.SugaredLogger
}

func (m *middleware) withLogging(h http.Handler) http.Handler {
	return logger.WithLogging(h, m.logger)
}

func (m *middleware) withCompressing(h http.Handler) http.Handler {
	return compress.WithCompressing(h, m.logger)
}

func Start() error {
	l, err := logger.Initialize()
	if err != nil {
		return err
	}
	c := config.GetConfig()
	s, err := storage.NewStorage(c)
	if err != nil {
		return err
	}
	h := handler{storage: s, config: c, logger: l}
	m := middleware{logger: l}

	router := chi.NewRouter()

	router.Use(m.withLogging)
	router.Use(m.withCompressing)

	router.Post("/", h.shortifyLink)
	router.Post("/api/shorten", h.shortenLink)
	router.Post("/api/shorten/batch", h.shortenLinkBatch)
	router.Get("/ping", h.checkDatabaseConnection)
	router.Get("/{hash}", h.getLink)

	l.Infow("Starting server", "address", c.ServerAddress)
	if err := http.ListenAndServe(c.ServerAddress, router); err != nil {
		return err
	}
	return nil
}
