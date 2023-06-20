package authorization

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const SECRET_KEY = "shouldBeSavedInEnvFile"

func WithAuthorization(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	authorizationMiddleware := func(w http.ResponseWriter, r *http.Request) {
		authToken, err := r.Cookie("AuthToken")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				token, err := generateJWTToken()
				if err != nil {
					logger.Errorw("Failed to get token string", "err", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				cookie := &http.Cookie{Name: "AuthToken", Value: token}
				http.SetCookie(w, cookie)
				h.ServeHTTP(w, r)
				return
			} else {
				logger.Errorw("Failed to get AuthToken", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		userId := GetUserIdFromJWTToken(authToken.Value)
		if userId == -1 {
			logger.Errorw("Failed to parse userId from jwt token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userId", userId)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(authorizationMiddleware)
}

func generateJWTToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: 1,
	})

	tokenString, err := token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", fmt.Errorf("failed to generate token string: %v", err)
	}

	return tokenString, nil

}

func GetUserIdFromJWTToken(tokenString string) int {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(SECRET_KEY), nil
		})

	if err != nil {
		return -1
	}

	if !token.Valid {
		return -1
	}

	return claims.UserID
}
