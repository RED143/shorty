package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

const contentEncodingKey = "Content-Encoding"
const encodingType = "gzip"

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return n, fmt.Errorf("failed to compress data: %w", err)
	}

	return n, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.ResponseWriter.Header().Set(contentEncodingKey, encodingType)
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	err := c.zw.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to init reader: %w", err)
	}

	return &compressReader{
		ReadCloser: r,
		zr:         zr,
	}, nil
}

func (c compressReader) Read(p []byte) (int, error) {
	n, err := c.zr.Read(p)
	if err != nil {
		return n, fmt.Errorf("failed to read data: %w", err)
	}

	return n, nil
}

func (c *compressReader) Close() error {
	if err := c.ReadCloser.Close(); err != nil {
		return fmt.Errorf("failed to close data: %w", err)
	}

	if err := c.zr.Close(); err != nil {
		return fmt.Errorf("failed to close gzip data: %w", err)
	}

	return nil
}

func WithCompressing(h http.Handler, logger *zap.SugaredLogger) http.Handler {
	compressMiddleware := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, encodingType)
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer func() {
				if err := cw.Close(); err != nil {
					logger.Errorf("Failed to close writer %v", err)
				}
			}()
		}

		contentEncoding := r.Header.Get(contentEncodingKey)
		sendsGzip := strings.Contains(contentEncoding, encodingType)
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logger.Errorf("Failed to compress data: %v", err)
				return
			}
			r.Body = cr
			defer func() {
				if err := cr.Close(); err != nil {
					logger.Errorf("Failed to close responser %v", err)
				}
			}()
		}

		h.ServeHTTP(ow, r)
	}

	return http.HandlerFunc(compressMiddleware)
}
