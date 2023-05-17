package handlers

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"shorty/internal/app/config"
	"shorty/internal/app/hash"
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
		log.Println("failed to parse request body")
		return
	}
	hashString := hash.Generate(body)
	str.Put(hashString, string(body))

	if len(body) == 0 {
		http.Error(writer, "URL should be provided", http.StatusBadRequest)
		return
	}

	fullURL, err := url.JoinPath(cfg.BaseAddress, hashString)
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		log.Println("failed to generate path")
		return
	}

	writer.Header().Set("content-type", "plain/text")
	writer.WriteHeader(http.StatusCreated)
	_, err = writer.Write([]byte(fullURL))
	if err != nil {
		http.Error(writer, "Internal server error", http.StatusInternalServerError)
		log.Println("failed to write response")
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
