package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func Initialize() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()

	if err != nil {
		return nil, err
	}
	defer logger.Sync()
	return logger.Sugar(), nil
}

func WithLogging(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		start := time.Now()
		h.ServeHTTP(&lw, r)
		duration := time.Since(start)

		logger.Infow("Request",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
		)

		logger.Infow("Response",
			"status", responseData.status,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
