package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"shorty/internal/app/authorization"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
	"shorty/internal/app/models"
	"shorty/internal/app/storage"
	"shorty/internal/app/storage/dbstorage"

	"go.uber.org/zap"
)

const internalServerError = "Internal server error"
const contentTypeKey = "content-type"
const applicationJSONType = "application/json"

func GetLink(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
	cfg config.Config,
	str storage.Storage,
	logger *zap.SugaredLogger,
) {
	shortURL, err := url.JoinPath(cfg.BaseAddress, request.URL.Path)
	if err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("failed to get shortURL: %v", err)
		return
	}
	link, err := str.Get(ctx, shortURL)
	if err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("failed to write response: %v", err)
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

func ShortenLink(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request, cfg config.Config,
	str storage.Storage,
	logger *zap.SugaredLogger,
) {
	var req models.ShortenRequest
	isJSONRequest := request.RequestURI == "/api/shorten"
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)

	if isJSONRequest {
		dec := json.NewDecoder(request.Body)
		if err := dec.Decode(&req); err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("cannot decode request JSON body: %v", err)
			return
		}
	} else {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("failed to parse request body: %v", err)
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
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("failed to generate shortURL: %v", err)
		return
	}
	err = str.Put(ctx, shortURL, req.URL, userID.(string))
	alreadySaved := errors.Is(err, dbstorage.ErrConflict)
	if err != nil && !alreadySaved {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("failed to save url: %v", err)
		return
	}

	statusCode := http.StatusCreated
	if alreadySaved {
		statusCode = http.StatusConflict
	}

	if isJSONRequest {
		resp := models.ShortenResponse{Result: shortURL}
		writer.Header().Set(contentTypeKey, applicationJSONType)
		writer.WriteHeader(statusCode)
		enc := json.NewEncoder(writer)
		if err := enc.Encode(resp); err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("error encoding response: %v", err)
			return
		}
		return
	}

	writer.Header().Set(contentTypeKey, "plain/text")
	writer.WriteHeader(statusCode)
	_, err = writer.Write([]byte(shortURL))
	if err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("failed to write response for shorten link: %v", err)
		return
	}
}

func ShortenLinkBatch(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
	cfg config.Config,
	str storage.Storage,
	logger *zap.SugaredLogger,
) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)

	var urls models.ShortenBatchRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&urls); err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("cannot decode request JSON body for batch request: %v", err)
		return
	}

	userURLs := make([]models.UserURLs, len(urls))
	for i, u := range urls {
		shortURL, err := hash.GenerateShortURL(u.OriginalURL, cfg.BaseAddress)
		if err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("failed to generate shortURL for batch request: %v", err)
			return
		}
		userURLs[i] = models.UserURLs{OriginalURL: u.OriginalURL, ShortURL: shortURL}
	}

	err := str.Batch(ctx, userURLs, userID.(string))
	if err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("Failed to batch saving: %v", err)
		return
	}

	var response models.ShortenBatchResponse
	for _, u := range urls {
		shortURL, err := hash.GenerateShortURL(u.OriginalURL, cfg.BaseAddress)
		if err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("failed to generate shortURL for batch response: %v", err)
			return
		}
		response = append(response, models.ShortenBatchResponseItem{CorrelationID: u.CorrelationID, ShortURL: shortURL})
	}

	writer.Header().Set(contentTypeKey, applicationJSONType)
	writer.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(response); err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("error encoding response for batch response: %v", err)
		return
	}
}

func CheckDatabaseConnection(
	ctx context.Context,
	writer http.ResponseWriter,
	str storage.Storage,
	logger *zap.SugaredLogger,
) {
	if err := str.Ping(ctx); err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("Failed to connect database: %v", err)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func GetUserURLs(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
	str storage.Storage,
	logger *zap.SugaredLogger,
) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)

	urls, err := str.UserURLs(ctx, userID.(string))
	if err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("failed to get user urls from storage: %v", err)
		return
	}

	writer.Header().Set(contentTypeKey, applicationJSONType)
	if len(urls) == 0 {
		writer.WriteHeader(http.StatusNoContent)
		return
	}

	writer.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(urls); err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("error encoding response for get user urls: %v", err)
		return
	}
}

func DeleteUserURLs(
	ctx context.Context,
	writer http.ResponseWriter,
	request *http.Request,
	cfg config.Config,
	str storage.Storage,
	logger *zap.SugaredLogger,
) {
	requestContext := request.Context()
	userID := requestContext.Value(authorization.UserIDContextKey)
	var req models.DeleteUrlsRequest

	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(writer, internalServerError, http.StatusInternalServerError)
		logger.Errorf("cannot decode request JSON body for deleting urls: %v", err)
		return
	}

	shortURLs := make([]string, len(req))
	for i, h := range req {
		shortURL, err := url.JoinPath(cfg.BaseAddress, h)
		if err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("failed to get shortURL for deleting urls: %v", err)
			return
		}

		shortURLs[i] = shortURL
	}

	writer.WriteHeader(http.StatusAccepted)
	go func() {
		err := deletingUserUrls(ctx, str, shortURLs, userID.(string))
		if err != nil {
			http.Error(writer, internalServerError, http.StatusInternalServerError)
			logger.Errorf("failed to delete URLs: %v", err)
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
			return fmt.Errorf("failed to delete user urls with userID=%s: %w", userID, err)
		}
		return nil
	}
}
