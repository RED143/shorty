package util

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func GetRouteID(request *http.Request, key string) string {
	return chi.URLParam(request, key)
}
