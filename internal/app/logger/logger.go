package logger

import (
	"log"
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

var sugar *zap.SugaredLogger

func Initialize() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("can't initialize logger: %v", err)
	}
	defer logger.Sync()
	sugar = logger.Sugar()
}

func Info(msg string, fields ...interface{}) {
	sugar.Infow(msg, fields...)
}

func WithLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		Info("Request",
			"uri", r.RequestURI,
			"method", r.Method,
			"duration", duration,
		)

		Info("Response",
			"status", responseData.status,
			"size", responseData.size,
		)

	}
}
