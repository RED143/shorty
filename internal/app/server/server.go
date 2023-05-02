package server

import (
	"net/http"

	"github.com/RED143/shorty/internal/app/handlers"
)

func routeHandler(writer http.ResponseWriter, request *http.Request) {
	switch request.URL.String() {
	case "/":
		handlers.ShortifyHandler(writer, request)
	default:
		handlers.LinkHandler(writer, request)
	}
}

func Start() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, routeHandler)

	err := http.ListenAndServe(`:8080`, mux)

	if err != nil {
		panic(err)
	}
}
