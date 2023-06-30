package handlers

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"shorty/internal/app/authorization"
	"shorty/internal/app/config"
	"shorty/internal/app/storage"
	"strings"
	"testing"
)

func TestShortenLink(t *testing.T) {
	configMock := config.Config{BaseAddress: "http://localhost:8080", ServerAddress: "localhost:8080", FileStoragePath: "", DatabaseDSN: ""}
	storageMock, err := storage.NewStorage(configMock)
	if err != nil {
		t.Errorf("failed to setup storage %v", err)
	}
	loggerMock := zaptest.NewLogger(t).Sugar()

	tests := []struct {
		name         string
		method       string
		expectedCode int
		contentType  string
		uri          string
		body         string
	}{
		{
			name:         "Should handle POST request with correct body",
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			contentType:  "text/plain",
			uri:          "/",
			body:         "www.google.com",
		},
		{
			name:         "Should handle POST request with correct json body",
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			contentType:  "application/json",
			uri:          "/shorten/api",
			body:         `{"url": "www.google.com"}`,
		},
		{
			name:         "Should return error for non-POST request",
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			contentType:  "text/plain",
			uri:          "/",
			body:         "",
		},
		{
			name:         "Should return error for empty body",
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			contentType:  "text/plain",
			uri:          "/",
			body:         "",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			ctx := context.WithValue(request.Context(), authorization.UserIDContextKey, "1")
			request.Header.Set("Content-Type", tc.contentType)
			writer := httptest.NewRecorder()

			ShortenLink(context.Background(), writer, request.WithContext(ctx), configMock, storageMock, loggerMock)

			assert.Equal(t, tc.expectedCode, writer.Code, "Got code %s; expected %s", writer.Code, tc.expectedCode)
		})
	}
}

func TestGetLink(t *testing.T) {
	configMock := config.Config{BaseAddress: "http://localhost:8080", ServerAddress: "localhost:8080", FileStoragePath: "", DatabaseDSN: ""}
	storageMock, err := storage.NewStorage(configMock)
	if err != nil {
		t.Errorf("failed to setup storage %v", err)
	}
	loggerMock := zaptest.NewLogger(t).Sugar()

	t.Run("Should return error for non-GET request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/asdf", nil)
		writer := httptest.NewRecorder()

		GetLink(context.Background(), writer, request, configMock, storageMock, loggerMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})

	t.Run("Should return error for link not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/asdf", nil)
		writer := httptest.NewRecorder()

		GetLink(context.Background(), writer, request, configMock, storageMock, loggerMock)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})
}
