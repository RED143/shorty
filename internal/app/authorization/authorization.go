package authorization

import (
	"fmt"
	"net/http"
)

func WithAuthorization(h http.Handler) http.Handler {
	authorizationMiddleware := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("authorizationMiddleware working")
	}

	return http.HandlerFunc(authorizationMiddleware)
}
