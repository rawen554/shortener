package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contentEncoding = "Content-Encoding"
	gzipValue       = "gzip"
)

type compressWriter struct {
	gin.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w gin.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
	}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	n, err := c.zw.Write(p)
	if err != nil {
		return 0, fmt.Errorf("error writing gzip: %w", err)
	}
	c.Header().Set("Content-Length", strconv.Itoa(n))

	return n, nil
}

func (c *compressWriter) WriteHeader(statusCode int) {
	c.Header().Set(contentEncoding, gzipValue)

	c.ResponseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	if err := c.zw.Close(); err != nil {
		return fmt.Errorf("error closing writer: %w", err)
	}
	return nil
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные.
type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("error creating gzip reader: %w", err)
	}

	return &compressReader{
		ReadCloser: r,
		zr:         zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	//nolint: wrapcheck // leads to unhandle errors in vendor packages
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.zr.Close(); err != nil {
		return fmt.Errorf("error closing gzip reader: %w", err)
	}

	return nil
}

func Compress() gin.HandlerFunc {
	return func(c *gin.Context) {
		ow := c.Writer

		acceptEncoding := c.Request.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, gzipValue)
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := newCompressWriter(c.Writer)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			//nolint: errcheck // leads to unhandled behaviour
			defer cw.Close()
		}

		contentEncoding := c.Request.Header.Get(contentEncoding)
		sendsGzip := strings.Contains(contentEncoding, gzipValue)
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(c.Request.Body)
			if err != nil {
				c.Writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			c.Request.Body = cr
			//nolint: errcheck // leads to unhandled behaviour
			defer cr.Close()
		}

		// передаём управление хендлеру
		c.Writer = ow
		c.Next()
	}
}
