package server

import (
	"context"
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

type handler struct {
	config  config.Config
	storage storage.Storage
	logger  *zap.SugaredLogger
	ctx     context.Context
}

func (h *handler) getLink(writer http.ResponseWriter, request *http.Request) {
	handlers.GetLink(h.ctx, writer, request, h.config, h.storage, h.logger)
}

func (h *handler) shortenLink(writer http.ResponseWriter, request *http.Request) {
	handlers.ShortenLink(h.ctx, writer, request, h.config, h.storage, h.logger)
}

func (h *handler) shortenLinkBatch(writer http.ResponseWriter, request *http.Request) {
	handlers.ShortenLinkBatch(h.ctx, writer, request, h.config, h.storage, h.logger)
}

func (h *handler) checkDatabaseConnection(writer http.ResponseWriter, request *http.Request) {
	handlers.CheckDatabaseConnection(h.ctx, writer, h.storage, h.logger)
}

func (h *handler) getUserURLs(writer http.ResponseWriter, request *http.Request) {
	handlers.GetUserURLs(h.ctx, writer, request, h.storage, h.logger)
}

func (h *handler) deleteUserURLs(writer http.ResponseWriter, request *http.Request) {
	handlers.DeleteUserURLs(h.ctx, writer, request, h.storage, h.logger)
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

func (m *middleware) withAuthorization(h http.Handler) http.Handler {
	return authorization.WithAuthorization(h, m.logger)
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
	defer s.Close()
	h := handler{storage: s, config: c, logger: l, ctx: context.Background()}
	m := middleware{logger: l}

	router := chi.NewRouter()

	router.Use(m.withLogging)
	router.Use(m.withAuthorization)
	router.Use(m.withCompressing)

	router.Post("/", h.shortenLink)
	router.Post("/api/shorten", h.shortenLink)
	router.Post("/api/shorten/batch", h.shortenLinkBatch)
	router.Get("/api/user/urls", h.getUserURLs)
	router.Delete("/api/user/urls", h.deleteUserURLs)
	router.Get("/ping", h.checkDatabaseConnection)
	router.Get("/{hash}", h.getLink)

	l.Infow("Starting server", "address", c.ServerAddress)
	if err := http.ListenAndServe(c.ServerAddress, router); err != nil {
		return err
	}
	return nil
}
