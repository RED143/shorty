package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortifyHandler(t *testing.T) {
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
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			writer := httptest.NewRecorder()

			ShortifyHandler(writer, request)

			assert.Equal(t, tc.expectedCode, writer.Code, "Got code %s; expected %s", writer.Code, tc.expectedCode)
		})
	}
}

func TestLinkHandler(t *testing.T) {
	t.Run("Should return error for non-GET request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/asdf", nil)
		writer := httptest.NewRecorder()

		LinkHandler(writer, request)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})

	t.Run("Should return error for link not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/asdf", nil)
		writer := httptest.NewRecorder()

		LinkHandler(writer, request)

		assert.Equal(t, http.StatusBadRequest, writer.Code, "Got code %s; expected %s", writer.Code, http.StatusBadRequest)
	})
}
