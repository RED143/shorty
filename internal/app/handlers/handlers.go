package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func GetLink(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	shortURL, err := url.JoinPath(cfg.BaseAddress, request.URL.Path)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to get shortURL", "err", err)
		return
	}
	link, err := str.Get(ctx, shortURL)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to write response", "err", err)
		return
	}

	if link.OriginalURL == "" {
		http.Error(writer, "Link not found", http.StatusBadRequest)
		return
	}

	if link.IsDeleted {
		writer.WriteHeader(http.StatusGone)
		return
	}

	writer.Header().Set("location", link.OriginalURL)
	writer.WriteHeader(http.StatusTemporaryRedirect)

}

func ShortenLink(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	var req models.ShortenRequest
	isJSONRequest := request.RequestURI == "/api/shorten"
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)

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

	shortURL, err := hash.GenerateShortURL(req.URL, cfg.BaseAddress)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to generate shortURL", "err", err)
		return
	}
	err = str.Put(ctx, shortURL, req.URL, userID.(string))
	alreadySaved := errors.Is(err, dbstorage.ErrConflict)
	if err != nil && !alreadySaved {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to save url", "err", err)
		return
	}

	statusCode := http.StatusCreated
	if alreadySaved {
		statusCode = http.StatusConflict
	}

	if isJSONRequest {
		resp := models.ShortenResponse{Result: shortURL}
		writer.Header().Set("content-type", "application/json")
		writer.WriteHeader(statusCode)
		enc := json.NewEncoder(writer)
		if err := enc.Encode(resp); err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("error encoding response", "err", err)
			return
		}
		return
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(statusCode)
	_, err = writer.Write([]byte(shortURL))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to write response", "err", err)
		return
	}

}

func ShortenLinkBatch(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)

	var urls models.ShortenBatchRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&urls); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("cannot decode request JSON body", "err", err)
		return
	}

	var userURLs []models.UserURLs
	for _, u := range urls {
		shortURL, err := hash.GenerateShortURL(u.OriginalURL, cfg.BaseAddress)
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to generate shortURL", "err", err)
			return
		}
		userURLs = append(userURLs, models.UserURLs{OriginalURL: u.OriginalURL, ShortURL: shortURL})
	}

	err := str.Batch(ctx, userURLs, userID.(string))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("Failed to batch saving", "err", err)
		return
	}

	var response models.ShortenBatchResponse
	for _, u := range urls {
		shortURL, err := hash.GenerateShortURL(u.OriginalURL, cfg.BaseAddress)
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to generate shortURL", "err", err)
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

func CheckDatabaseConnection(ctx context.Context, writer http.ResponseWriter, str storage.Storage, logger *zap.SugaredLogger) {
	if err := str.Ping(ctx); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("Failed to connect database", "err", err)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func GetUserURLs(ctx context.Context, writer http.ResponseWriter, request *http.Request, str storage.Storage, logger *zap.SugaredLogger) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)

	urls, err := str.UserURLs(ctx, userID.(string))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("failed to get user urls", "err", err)
		return
	}

	writer.Header().Set("content-type", "application/json")
	if len(urls) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	writer.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(urls); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("error encoding response", "err", err)
		return
	}
}

func DeleteUserURLs(ctx context.Context, writer http.ResponseWriter, request *http.Request, cfg config.Config, str storage.Storage, logger *zap.SugaredLogger) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)
	var req models.DeleteUrlsRequest

	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Errorw("cannot decode request JSON body", "err", err)
		return
	}

	var shortURLs []string

	for _, h := range req {
		shortURL, err := url.JoinPath(cfg.BaseAddress, h)
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to get shortURL", "err", err)
			return
		}

		shortURLs = append(shortURLs, shortURL)
	}

	writer.WriteHeader(http.StatusAccepted)
	go func() {
		err := deletingUserUrls(ctx, str, shortURLs, userID.(string))
		if err != nil {
			http.Error(writer, "Internal server error", http.StatusInternalServerError)
			logger.Errorw("failed to delete URLs", "err", err)
			return
		}
	}()
}

func deletingUserUrls(ctx context.Context, str storage.Storage, urls []string, userID string) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("deleting was interrupted by context for userID=%s", userID)
	default:
		if err := str.DeleteUserURls(ctx, urls, userID); err != nil {
			return fmt.Errorf("failed to delete user urls with userID=%s: %v", userID, err)
		}
		return nil
	}
}
