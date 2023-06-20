package authorization

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type ContextKey string

const secret = "shouldBeSavedInEnvFile"

func WithAuthorization(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	authorizationMiddleware := func(w http.ResponseWriter, r *http.Request) {
		authToken, err := r.Cookie("AuthToken")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				id := uuid.NewString()
				token, err := generateJWTToken(id)
				if err != nil {
					logger.Errorw("Failed to get token string", "err", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				ctx := context.WithValue(r.Context(), ContextKey("userID"), id)
				setAuthCookie(w, token)
				h.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				logger.Errorw("Failed to get AuthToken", "err", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		userID := GetUserIDFromJWTToken(authToken.Value)
		if userID == "" {
			logger.Errorw("Failed to parse userID from jwt token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextKey("userID"), userID)
		setAuthCookie(w, authToken.Value)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(authorizationMiddleware)
}

func generateJWTToken(id string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: id,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token string: %v", err)
	}

	return tokenString, nil

}

func GetUserIDFromJWTToken(tokenString string) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})

	if err != nil {
		return ""
	}

	if !token.Valid {
		return ""
	}

	return claims.UserID
}

func setAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{Name: "AuthToken", Value: token}
	http.SetCookie(w, cookie)
}
