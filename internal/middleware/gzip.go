package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/coolycow/gophermart/internal/error"
	"github.com/coolycow/gophermart/internal/logger"
	"github.com/gin-gonic/gin"
)

// RequestGzip — middleware для распаковки сжатых данных запроса
func RequestGzip() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.GetHeader("Content-Encoding"), "gzip") {
			logger.Log.Info("request with gzip")

			compressedData, _ := io.ReadAll(c.Request.Body)
			reader, err := gzip.NewReader(bytes.NewReader(compressedData))

			if err != nil {
				_ = c.Error(error.HTTPError{
					Message:    "Failed to decompress gzip data",
					StatusCode: http.StatusBadRequest,
				})
				c.Abort()
			}

			defer func(reader *gzip.Reader) {
				if reader != nil {
					err = reader.Close()
					if err != nil {
						logger.Log.Error(err.Error())
					}
				}
			}(reader)

			if reader != nil {
				c.Request.Body = io.NopCloser(reader)
			}
		}

		// Передаём управление другим Middleware или Handler
		c.Next()
	}
}
