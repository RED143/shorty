package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"shorty/internal/app/storage"
	"shorty/internal/app/storage/dbstorage"
)

func Shortify(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
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
	err = str.Put(ctx, hashString, string(body))
	alreadySaved := errors.Is(err, dbstorage.ErrConflict)
	if err != nil && !alreadySaved {
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

	statusCode := http.StatusCreated
	if alreadySaved {
		statusCode = http.StatusConflict
	}
	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(statusCode)
	_, err = writer.Write([]byte(fullURL))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to write response", "err", err)
		return
	}
}

func GetLink(ctx context.Context, writer http.ResponseWriter, request *http.Request, hash string, str storage.Storage, logger *zap.SugaredLogger) {
	link, err := str.Get(ctx, hash)
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

func ShortenLink(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
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
	err := str.Put(ctx, hashString, req.URL)
	alreadySaved := errors.Is(err, dbstorage.ErrConflict)
	if err != nil && !alreadySaved {
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

	statusCode := http.StatusCreated
	if alreadySaved {
		statusCode = http.StatusConflict
	}
	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(statusCode)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(resp); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("error encoding response", "err", err)
		return
	}
}

func ShortenLinkBatch(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	var urls models.ShortenBatchRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&urls); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("cannot decode request JSON body", "err", err)
		return
	}

	err := str.Batch(ctx, urls)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("Failed to batch saving", "err", err)
		return
	}

	var response models.ShortenBatchResponse
	for _, u := range urls {
		shortURL, err := url.JoinPath(cfg.BaseAddress, hash.Generate([]byte(u.OriginalURL)))
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to generate path", "err", err)
			return
		}
		response = append(response, models.ShortenBatchResponseItem{CorrelationID: u.CorrelationID, ShortURL: shortURL})
	}

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(response); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("error encoding response", "err", err)
		return
	}
}

func CheckDatabaseConnection(ctx context.Context, writer http.ResponseWriter, request *http.Request, str storage.Storage, logger *zap.SugaredLogger) {
	if err := str.Ping(ctx); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("Failed to connect database", "err", err)
		return
	}
	writer.WriteHeader(http.StatusOK)
}
