package server

import (
	"fmt"
	"net/http"
	"shorty/internal/app/authorization"
	"shorty/internal/app/compress"
	"shorty/internal/app/config"
	"shorty/internal/app/handlers"
	"shorty/internal/app/logger"
	"shorty/internal/app/storage"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const userUrlsPath = "/api/user/urls"

type handler struct {
	logger  *zap.SugaredLogger
	storage storage.Storage
	config  config.Config
}

func (h *handler) getLink(writer http.ResponseWriter, request *http.Request) {
	handlers.GetLink(request.Context(), writer, request, h.config, h.storage, h.logger)
}

func (h *handler) shortenLink(writer http.ResponseWriter, request *http.Request) {
	handlers.ShortenLink(request.Context(), writer, request, h.config, h.storage, h.logger)
}

func (h *handler) shortenLinkBatch(writer http.ResponseWriter, request *http.Request) {
	handlers.ShortenLinkBatch(request.Context(), writer, request, h.config, h.storage, h.logger)
}

func (h *handler) checkDatabaseConnection(writer http.ResponseWriter, request *http.Request) {
	handlers.CheckDatabaseConnection(request.Context(), writer, h.storage, h.logger)
}

func (h *handler) getUserURLs(writer http.ResponseWriter, request *http.Request) {
	handlers.GetUserURLs(request.Context(), writer, request, h.storage, h.logger)
}

func (h *handler) deleteUserURLs(writer http.ResponseWriter, request *http.Request) {
	handlers.DeleteUserURLs(request.Context(), writer, request, h.config, h.storage, h.logger)
}

type middleware struct {
	logger *zap.SugaredLogger
	cfg    config.Config
}

func (m *middleware) withLogging(h http.Handler) http.Handler {
	return logger.WithLogging(h, m.logger)
}

func (m *middleware) withCompressing(h http.Handler) http.Handler {
	return compress.WithCompressing(h, m.logger)
}

func (m *middleware) withAuthorization(h http.Handler) http.Handler {
	return authorization.WithAuthorization(h, m.cfg, m.logger)
}

func Start() error {
	l, err := logger.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	c := config.GetConfig()
	s, err := storage.NewStorage(c)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			l.Errorf("failed to close storage: %v", err)
		}
	}()

	h := handler{storage: s, config: c, logger: l}
	m := middleware{logger: l, cfg: c}

	router := chi.NewRouter()

	router.Use(m.withLogging)
	router.Use(m.withAuthorization)
	router.Use(m.withCompressing)

	router.Post("/", h.shortenLink)
	router.Post("/api/shorten", h.shortenLink)
	router.Post("/api/shorten/batch", h.shortenLinkBatch)
	router.Get(userUrlsPath, h.getUserURLs)
	router.Delete(userUrlsPath, h.deleteUserURLs)
	router.Get("/ping", h.checkDatabaseConnection)
	router.Get("/{hash}", h.getLink)

	l.Infof("Starting server on address: %s", c.ServerAddress)
	if err := http.ListenAndServe(c.ServerAddress, router); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
