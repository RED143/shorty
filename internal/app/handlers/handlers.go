package handlers

import (
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
	}

	body, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "Failed to parse request body", http.StatusInternalServerError)
	}
	hashString := hash.Generate(body)
	storage.SetValue(hashString, string(body))

	if len(body) == 0 {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
	}

	cfg := config.GetConfig()
	fullURL, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		http.Error(writer, "Failed to parse request body", http.StatusInternalServerError)
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(fullURL))
}

func GetLink(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Only GET requests are allowed", http.StatusBadRequest)
	}

	hash := chi.URLParam(request, "hash")
	link, ok := storage.GetValue(hash)

	if !ok {
		http.Error(writer, "Link not found", http.StatusBadRequest)
	}

	writer.Header().Set("location", link)
	writer.WriteHeader(http.StatusTemporaryRedirect)
}
