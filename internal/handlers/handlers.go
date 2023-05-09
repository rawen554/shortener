package handlers

import (
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/flags"
	"github.com/rawen554/shortener/internal/utils"
)

func RedirectToOriginal(urls map[string][]byte) func(c *gin.Context) {
	return func(c *gin.Context) {
		res := c.Writer
		id := c.Param("id")
		originalURL := urls[id]
		if originalURL == nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Location", string(originalURL))
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func ShortenURL(urls map[string][]byte) func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.Request
		res := c.Writer
		body, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		id, err := utils.GenerateRandomString(8)
		if err != nil {
			panic(err)
		}
		urls[id] = body

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		result, err := url.JoinPath(flags.Config.RedirectBaseURL, id)
		if err != nil {
			panic(err)
		}
		res.Write([]byte(result))
	}
}
