package authorization

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"shorty/internal/app/config"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

type ContextKey string

const UserIDContextKey = ContextKey("userID")

func WithAuthorization(h http.Handler, cfg config.Config, logger *zap.SugaredLogger) http.Handler {
	authorizationMiddleware := func(w http.ResponseWriter, r *http.Request) {
		authToken, err := r.Cookie("AuthToken")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				id := uuid.NewString()
				token, err := generateJWTToken(id, cfg)
				if err != nil {
					logger.Errorf("Failed to get token string: %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				ctx := context.WithValue(r.Context(), UserIDContextKey, id)
				setAuthCookie(w, token)
				h.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			logger.Errorf("Failed to get AuthToken: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		userID := GetUserIDFromJWTToken(authToken.Value, cfg)
		if userID == "" {
			logger.Errorw("Failed to parse userID from jwt token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		setAuthCookie(w, authToken.Value)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(authorizationMiddleware)
}

func generateJWTToken(id string, cfg config.Config) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: id,
	})

	tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token string: %w", err)
	}

	return tokenString, nil
}

func GetUserIDFromJWTToken(tokenString string, cfg config.Config) string {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
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
