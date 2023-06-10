package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"shorty/internal/app/storage"
	"time"
)

func Shortify(writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to parse request body", "err", err)
		return
	}

	if len(body) == 0 {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	hashString := hash.Generate(body)
	if err := str.Put(hashString, string(body)); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to save url", "err", err)
		return
	}

	fullURL, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to generate path", "err", err)
		return
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte(fullURL))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to write response", "err", err)
		return
	}
}

func GetLink(writer http.ResponseWriter, request *http.Request, hash string, str storage.Storage, logger *zap.SugaredLogger) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Only GET requests are allowed", http.StatusBadRequest)
		return
	}

	link, err := str.Get(hash)

	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to write response", "err", err)
		return
	}

	if link == "" {
		http.Error(writer, "Link not found", http.StatusBadRequest)
		return
	}

	writer.Header().Set("location", link)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func ShortenLink(writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	var req models.ShortenRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("cannot decode request JSON body", "err", err)
		return
	}

	if req.URL == "" {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	hashString := hash.Generate([]byte(req.URL))
	err := str.Put(hashString, req.URL)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to save url", "err", err)
		return
	}

	result, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to generate path", "err", err)
		return
	}

	resp := models.ShortenResponse{Result: result}

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(resp); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("error encoding response", "err", err)
		return
	}
}

func CheckDatabaseConnection(writer http.ResponseWriter, request *http.Request, db *sql.DB, logger *zap.SugaredLogger) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Only GET requests are allowed", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("Failed to connect database", "err", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
}
