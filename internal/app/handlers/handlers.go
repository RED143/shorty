package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"shorty/internal/app/authorization"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"shorty/internal/app/storage"
	"shorty/internal/app/storage/dbstorage"
)

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
	isJSONRequest := request.RequestURI == "/api/shorten"
	requestContext := request.Context()
	userID := requestContext.Value(authorization.ContextKey("userID"))

	if isJSONRequest {
		dec := json.NewDecoder(request.Body)
		if err := dec.Decode(&req); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("cannot decode request JSON body", "err", err)
			return
		}
	} else {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to parse request body", "err", err)
			return
		}
		req.URL = string(body)
	}

	if req.URL == "" {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	hashString := hash.Generate([]byte(req.URL))
	err := str.Put(ctx, hashString, req.URL, userID.(string))
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

	statusCode := http.StatusCreated
	if alreadySaved {
		statusCode = http.StatusConflict
	}

	if isJSONRequest {
		resp := models.ShortenResponse{Result: result}
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(statusCode)
		enc := json.NewEncoder(writer)
		if err := enc.Encode(resp); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("error encoding response", "err", err)
			return
		}
	} else {
		writer.Header().Set("content-type", "plain/text")
		writer.WriteHeader(statusCode)
		_, err = writer.Write([]byte(result))
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to write response", "err", err)
			return
		}
	}

}

func ShortenLinkBatch(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.ContextKey("userID"))

	var urls models.ShortenBatchRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&urls); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("cannot decode request JSON body", "err", err)
		return
	}

	err := str.Batch(ctx, urls, userID.(string))
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

func GetUserURLs(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.ContextKey("userID"))

	urls, err := str.UserURLs(ctx, userID.(string))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to get user urls", "err", err)
		return
	}

	var response []models.UserURL

	for _, u := range urls {
		shortURL, err := url.JoinPath(cfg.BaseAddress, u.Hash)
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to generate path", "err", err)
			return
		}
		response = append(response, models.UserURL{ShortURL: shortURL, OriginalURL: u.URL})
	}

	writer.Header().Set("content-type", "application/json")
	if len(response) == 0 {
		writer.WriteHeader(http.StatusNoContent)
	} else {
		writer.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(writer)
		if err := enc.Encode(response); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("error encoding response", "err", err)
			return
		}
	}
}
