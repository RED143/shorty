package handlers

import (
	"io"
	"net/http"

	"github.com/RED143/shorty/internal/app/hash"
	"github.com/RED143/shorty/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func ShortifyHandler(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Only POST requests are allowed", http.StatusBadRequest)
		return
	}

	body, _ := io.ReadAll(request.Body)
	hashString := hash.Generate(body)
	storage.SetValue(hashString, string(body))

	if len(body) == 0 {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte("http://localhost:8080/" + hashString[:7]))
}

func LinkHandler(writer http.ResponseWriter, request *http.Request) {
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
