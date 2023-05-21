package handlers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"shorty/internal/app/config"
	"shorty/internal/app/storage"
	"strings"
	"testing"
)

func TestShortify(t *testing.T) {
	configMock := config.Config{BaseAddress: "http://localhost:8080", ServerAddress: "localhost:8080"}
	storageMock := storage.NewStorage()

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
	storageMock := storage.NewStorage()
	hash := "asdf"
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
	storageMock := storage.NewStorage()

	t.Run("Should return error for non-POST request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/shorten", nil)
		writer := httptest.NewRecorder()

		ShortenLink(writer, request, configMock, storageMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})
}
