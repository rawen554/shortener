package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/app"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var ErrInitStore = errors.New("failed to initialize a new storage")
var ErrSaveData = errors.New("failed to save data")
var ErrDeleteDataFile = errors.New("failed to delete data file")
var ErrClosingBody = errors.New("failed to close body")

const (
	tmpJSONPath = "./test.json"
	contentType = "Content-Type"
)

func Test_redirectToOriginal(t *testing.T) {
	type args struct {
		urls           map[string]string
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
				urls: map[string]string{
					"1": "http://ya.ru",
				},
				originalURL:    "http://ya.ru",
				shortURL:       "/1",
				shouldRedirect: true,
			},
		},
		{
			name: "error short url not found",
			args: args{
				urls: map[string]string{
					"1": "http://ya.ru",
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

			storage, err := fs.NewFileStorage(tmpJSONPath)
			if err != nil {
				t.Error(errors.Join(ErrInitStore, err))
				return
			}
			defer func() {
				err := storage.DeleteStorageFile()
				if err != nil {
					t.Error(errors.Join(ErrDeleteDataFile, err))
					return
				}
			}()

			for url := range tt.args.urls {
				_, err := storage.Put(url, tt.args.urls[url], "")
				if err != nil {
					t.Error(errors.Join(ErrSaveData, err))
					return
				}
			}

			testApp := app.NewApp(&config.ServerConfig{}, storage)
			r := setupRouter(testApp)
			req := httptest.NewRequest(http.MethodGet, tt.args.shortURL, nil)

			r.ServeHTTP(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Error(errors.Join(ErrClosingBody, err))
					return
				}
			}()

			if tt.args.shouldRedirect {
				assert.Equal(t, tt.args.originalURL, res.Header.Get("Location"))
				assert.Equal(t, http.StatusTemporaryRedirect, res.StatusCode)
			} else {
				assert.Equal(t, http.StatusNotFound, res.StatusCode)
			}
		})
	}
}

func Test_shortURL_V1(t *testing.T) {
	type args struct {
		urls        map[string]string
		originalURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add new url to empty map",
			args: args{
				urls:        make(map[string]string),
				originalURL: "https://ya.ru",
			},
		},
		{
			name: "add new url to map",
			args: args{
				urls: map[string]string{
					"abc": "https://ya.com",
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

			storage, err := fs.NewFileStorage(tmpJSONPath)
			if err != nil {
				t.Error(errors.Join(ErrInitStore, err))
				return
			}
			defer func() {
				err := storage.DeleteStorageFile()
				if err != nil {
					t.Error(errors.Join(ErrDeleteDataFile, err))
					return
				}
			}()

			for url := range tt.args.urls {
				_, err := storage.Put(url, tt.args.urls[url], "")
				if err != nil {
					t.Error(errors.Join(ErrSaveData, err))
					return
				}
			}

			testApp := app.NewApp(&config.ServerConfig{}, storage)
			r := setupRouter(testApp)
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(tt.args.originalURL)))
			req.Header.Add(contentType, "text/plain")

			r.ServeHTTP(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Errorf("error closing body: %v", err)
					return
				}
			}()
			body, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}

func Test_shortURL_V2(t *testing.T) {
	type args struct {
		urls        map[string]string
		originalURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "add new url to empty map",
			args: args{
				urls:        make(map[string]string),
				originalURL: "https://ya.ru",
			},
		},
		{
			name: "add new url to map",
			args: args{
				urls: map[string]string{
					"abc": "https://ya.com",
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

			storage, err := fs.NewFileStorage(tmpJSONPath)
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer func() {
				err := storage.DeleteStorageFile()
				if err != nil {
					t.Errorf("failed to delete data: %v", err)
					return
				}
			}()

			for url := range tt.args.urls {
				_, err := storage.Put(url, tt.args.urls[url], "")
				if err != nil {
					t.Errorf("failed to save data: %v", err)
					return
				}
			}

			testApp := app.NewApp(&config.ServerConfig{}, storage)
			r := setupRouter(testApp)
			reqObj := models.ShortenReq{
				URL: tt.args.originalURL,
			}
			obj, err := json.Marshal(reqObj)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(obj))
			req.Header.Add("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			res := w.Result()
			defer func() {
				err := res.Body.Close()
				if err != nil {
					t.Errorf("failed to close body: %v", err)
					return
				}
			}()
			body, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}
