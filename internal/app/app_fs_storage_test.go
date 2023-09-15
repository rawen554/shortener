package app

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testConfig = &config.ServerConfig{
	RunAddr: ":8080",
	Secret:  "b4952c3809196592c026529df00774e46bfb5be0",
}

const (
	ErrorDeletingTestFile = "error deleting test file: %v"
	ErrorStoringRecord    = "error storing value: %v"
	ErrorClosingBody      = "error close body: %v"
	ErrorSetupRouter      = "failed to setup router: %v"
	ErrorSetupStorage     = "failed to initialize a new storage: %v"

	TestStoragePath = "./test.json"
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

			storage, err := fs.NewFileStorage(TestStoragePath)
			if err != nil {
				t.Errorf(ErrorSetupStorage, err)
				return
			}
			defer func() {
				if err := storage.DeleteStorageFile(); err != nil {
					t.Errorf(ErrorDeletingTestFile, err)
				}
			}()

			for url := range tt.args.urls {
				if _, err := storage.Put(url, tt.args.urls[url], ""); err != nil {
					t.Errorf(ErrorStoringRecord, err)
				}
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf(ErrorSetupRouter, err)
			}
			req := httptest.NewRequest(http.MethodGet, tt.args.shortURL, nil)

			r.ServeHTTP(w, req)

			res := w.Result()
			if err := res.Body.Close(); err != nil {
				t.Errorf(ErrorClosingBody, err)
			}

			if tt.args.shouldRedirect {
				assert.Equal(t, tt.args.originalURL, res.Header.Get(location))
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

			storage, err := fs.NewFileStorage(TestStoragePath)
			if err != nil {
				t.Errorf(ErrorSetupStorage, err)
				return
			}
			defer func() {
				if err := storage.DeleteStorageFile(); err != nil {
					t.Errorf(ErrorDeletingTestFile, err)
				}
			}()

			for url := range tt.args.urls {
				if _, err := storage.Put(url, tt.args.urls[url], ""); err != nil {
					t.Errorf(ErrorStoringRecord, err)
				}
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf(ErrorSetupRouter, err)
			}
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(tt.args.originalURL)))
			req.Header.Add(contentType, textPlain)

			r.ServeHTTP(w, req)

			res := w.Result()
			if err := res.Body.Close(); err != nil {
				t.Errorf(ErrorClosingBody, err)
			}
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

			storage, err := fs.NewFileStorage(TestStoragePath)
			if err != nil {
				t.Errorf(ErrorSetupStorage, err)
				return
			}
			defer func() {
				if err := storage.DeleteStorageFile(); err != nil {
					t.Errorf(ErrorDeletingTestFile, err)
				}
			}()

			for url := range tt.args.urls {
				if _, err := storage.Put(url, tt.args.urls[url], ""); err != nil {
					t.Errorf(ErrorStoringRecord, err)
				}
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf(ErrorSetupRouter, err)
			}
			reqObj := models.ShortenReq{
				URL: tt.args.originalURL,
			}
			obj, err := json.Marshal(reqObj)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, apiShortenPath, bytes.NewBuffer(obj))
			req.Header.Add(contentType, applicationJSON)

			r.ServeHTTP(w, req)

			res := w.Result()
			if err := res.Body.Close(); err != nil {
				t.Errorf(ErrorClosingBody, err)
			}
			body, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}
