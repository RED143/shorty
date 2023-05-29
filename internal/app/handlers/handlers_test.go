package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"shorty/internal/app/config"
	"shorty/internal/app/models"
	"shorty/internal/app/storage"
	"strings"
	"testing"
)

func getStoragePath(t *testing.T) string {
	dir, err := os.Getwd()
	if err != nil {
		t.Error("failed to get current path")
	}
	return filepath.Join(dir, "test.json")
}

func deleteStorageFile(t *testing.T) {
	os.Remove(getStoragePath(t))
}

func TestShortify(t *testing.T) {
	configMock := config.Config{BaseAddress: "http://localhost:8080", ServerAddress: "localhost:8080"}
	storageMock := storage.NewStorage(getStoragePath(t))
	defer deleteStorageFile(t)

	tests := []struct {
		name         string
		method       string
		expectedCode int
		body         string
	}{
		{
			name:         "Should handle POST request with correct body",
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			body:         "www.google.com",
		},
		{
			name:         "Should return error for non-POST request",
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			body:         "",
		},
		{
			name:         "Should return error for empty body",
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			body:         "",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			writer := httptest.NewRecorder()

			Shortify(writer, request, configMock, storageMock)

			assert.Equal(t, tc.expectedCode, writer.Code, "Got code %s; expected %s", writer.Code, tc.expectedCode)
		})
	}
}

func TestGetLink(t *testing.T) {
	hash := "asdf"
	storageMock := storage.NewStorage(getStoragePath(t))
	defer deleteStorageFile(t)

	t.Run("Should return error for non-GET request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/asdf", nil)
		writer := httptest.NewRecorder()

		GetLink(writer, request, hash, storageMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})

	t.Run("Should return error for link not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/asdf", nil)
		writer := httptest.NewRecorder()

		GetLink(writer, request, hash, storageMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})
}

func TestShortenLink(t *testing.T) {
	configMock := config.Config{BaseAddress: "http://localhost:8080", ServerAddress: "localhost:8080"}
	storageMock := storage.NewStorage(getStoragePath(t))
	defer deleteStorageFile(t)

	t.Run("Should return error for non-POST request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/shorten", nil)
		writer := httptest.NewRecorder()

		ShortenLink(writer, request, configMock, storageMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})

	t.Run("Should return error if url not provided", func(t *testing.T) {
		data := models.ShortenRequest{URL: ""}
		reqData, _ := json.Marshal(data)
		request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(reqData))
		writer := httptest.NewRecorder()

		ShortenLink(writer, request, configMock, storageMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})
}
