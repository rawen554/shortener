package handlers

import (
	"io"
	"net/http"
	"net/url"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/flags"
	"github.com/rawen554/shortener/internal/utils"
)

type SyncURLMap struct {
	mux  *sync.Mutex
	Urls map[string][]byte
}

func NewSyncURLMap(urls map[string][]byte) SyncURLMap {
	return SyncURLMap{
		Urls: urls,
		mux:  &sync.Mutex{},
	}
}

func RedirectToOriginal(m SyncURLMap) func(c *gin.Context) {
	return func(c *gin.Context) {
		res := c.Writer
		id := c.Param("id")

		m.mux.Lock()
		originalURL := m.Urls[id]
		m.mux.Unlock()

		if originalURL == nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Location", string(originalURL))
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func ShortenURL(m SyncURLMap) func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.Request
		res := c.Writer
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		id, err := utils.GenerateRandomString(8)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		result, err := url.JoinPath(flags.Config.RedirectBaseURL, id)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		m.mux.Lock()
		m.Urls[id] = body
		m.mux.Unlock()

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(result))
	}
}
