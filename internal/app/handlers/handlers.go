package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
	"shorty/internal/app/storage"
)

func Shortify(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		fmt.Errorf("failed to parse body: %w", err)
	}
	hashString := hash.Generate(body)
	storage.SetValue(hashString, string(body))

	if len(body) == 0 {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	cfg := config.GetConfig()
	fullUrl, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		fmt.Errorf("failed to joinPath(%s, %s): %w", cfg.BaseAddress, hashString, err)
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(fullUrl))
}

func GetLink(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Only GET requests are allowed", http.StatusBadRequest)
		return
	}

	hash := chi.URLParam(request, "hash")
	link := storage.GetValue(hash)

	if link == "" {
		http.Error(writer, "Link not found", http.StatusBadRequest)
		return
	}

	writer.Header().Set("location", link)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}
