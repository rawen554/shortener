package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_redirectToOriginal(t *testing.T) {
	type args struct {
		urls           map[string][]byte
		shortURL       string
		originalURL    string
		shouldRedirect bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "simple redirect",
			args: args{
				urls: map[string][]byte{
					"1": []byte("http://ya.ru"),
				},
				originalURL:    "http://ya.ru",
				shortURL:       "/1",
				shouldRedirect: true,
			},
		},
		{
			name: "error short url not found",
			args: args{
				urls: map[string][]byte{
					"1": []byte("http://ya.ru"),
				},
				originalURL:    "http://ya.ru",
				shortURL:       "/2",
				shouldRedirect: false,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			app := app.NewApp(&config.ServerConfig{}, store.NewStorage(tt.args.urls))
			r := setupRouter(app)
			req := httptest.NewRequest(http.MethodGet, tt.args.shortURL, nil)

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if tt.args.shouldRedirect {
				assert.Equal(t, tt.args.originalURL, res.Header.Get("Location"))
				assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
			} else {
				assert.Equal(t, http.StatusNotFound, res.StatusCode)
			}
		})
	}
}

func Test_shortURL(t *testing.T) {
	type args struct {
		urls        map[string][]byte
		originalURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add new url to empty map",
			args: args{
				urls:        make(map[string][]byte),
				originalURL: "https://ya.ru",
			},
		},
		{
			name: "add new url to map",
			args: args{
				urls: map[string][]byte{
					"abc": []byte("https://ya.com"),
				},
				originalURL: "https://ya.ru",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			app := app.NewApp(&config.ServerConfig{}, store.NewStorage(tt.args.urls))
			r := setupRouter(app)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(tt.args.originalURL)))

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}
