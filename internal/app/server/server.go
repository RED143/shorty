package server

import (
	"database/sql"
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
	storage *storage.Storage
	logger  *zap.SugaredLogger
	db      *sql.DB
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

func (h *handler) checkDatabaseConnection(writer http.ResponseWriter, request *http.Request) {
	handlers.CheckDatabaseConnection(writer, request, h.db, h.logger)
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
	logger, err := logger.Initialize()
	if err != nil {
		return err
	}
	c := config.GetConfig()
	s, err := storage.NewStorage(c.FileStoragePath)
	if err != nil {
		return err
	}
	db, err := sql.Open("pgx", c.DatabaseDSN)
	if err != nil {
		return err
	}
	defer db.Close()
	h := handler{storage: s, config: c, logger: logger, db: db}
	m := middleware{logger: logger}

	router := chi.NewRouter()

	router.Use(m.withLogging)
	router.Use(m.withCompressing)

	router.Post("/", h.shortifyLink)
	router.Post("/api/shorten", h.shortenLink)
	router.Get("/ping", h.checkDatabaseConnection)
	router.Get("/{hash}", h.getLink)

	logger.Infow("Starting server", "address", c.ServerAddress)
	if err := http.ListenAndServe(c.ServerAddress, router); err != nil {
		return err
	}
	return nil
}
