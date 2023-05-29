package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
	"shorty/internal/app/logger"
	"shorty/internal/app/models"
	"shorty/internal/app/storage"
)

func Shortify(writer http.ResponseWriter, request *http.Request, cfg config.Config, str *storage.Storage) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Debug("failed to parse request body", "err", err)
		return
	}

	if len(body) == 0 {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	hashString := hash.Generate(body)
	str.Put(hashString, string(body))

	fullURL, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Debug("failed to generate path", "err", err)
		return
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte(fullURL))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Debug("failed to write response", "err", err)
		return
	}
}

func GetLink(writer http.ResponseWriter, request *http.Request, hash string, str *storage.Storage) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Only GET requests are allowed", http.StatusBadRequest)
		return
	}

	link, ok := str.Get(hash)

	if !ok {
		http.Error(writer, "Link not found", http.StatusBadRequest)
		return
	}

	writer.Header().Set("location", link)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}

func ShortenLink(writer http.ResponseWriter, request *http.Request, cfg config.Config, str *storage.Storage) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	var req models.ShortenRequest
	dec := json.NewDecoder(request.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Debug("cannot decode request JSON body", "err", err)
		return
	}

	if req.URL == "" {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	hashString := hash.Generate([]byte(req.URL))
	str.Put(hashString, req.URL)

	result, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Debug("failed to generate path", "err", err)
		return
	}

	resp := models.ShortenResponse{Result: result}

	writer.Header().Set("content-type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(writer)
	if err := enc.Encode(resp); err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		logger.Debug("error encoding response", "err", err)
		return
	}
}
