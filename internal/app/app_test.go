package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/middleware/auth"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/rawen554/shortener/internal/store/mocks"
	"github.com/rawen554/shortener/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var testConfig = &config.ServerConfig{
	RunAddr: ":8080",
	Secret:  "b4952c3809196592c026529df00774e46bfb5be0",
}

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

			storage, err := fs.NewFileStorage("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer storage.DeleteStorageFile()

			for url := range tt.args.urls {
				storage.Put(url, tt.args.urls[url], "")
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf("failed to setup router: %v", err)
			}
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

			storage, err := fs.NewFileStorage("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer storage.DeleteStorageFile()

			for url := range tt.args.urls {
				storage.Put(url, tt.args.urls[url], "")
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf("failed to setup router: %v", err)
			}
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer([]byte(tt.args.originalURL)))
			req.Header.Add("Content-Type", "text/plain")

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
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

			storage, err := fs.NewFileStorage("./test.json")
			if err != nil {
				t.Errorf("failed to initialize a new storage: %v", err)
				return
			}
			defer storage.DeleteStorageFile()

			for url := range tt.args.urls {
				storage.Put(url, tt.args.urls[url], "")
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf("failed to setup router: %v", err)
			}
			reqObj := models.ShortenReq{
				URL: tt.args.originalURL,
			}
			obj, err := json.Marshal(reqObj)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(obj))
			req.Header.Add("Content-Type", "application/json")

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.NotEmpty(t, body)
		})
	}
}

func BenchmarkShortUrl(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	length := 10

	storage, err := fs.NewFileStorage("./test.json")
	if err != nil {
		b.Errorf("failed to initialize a new storage: %v", err)
		return
	}
	defer storage.DeleteStorageFile()

	testApp := NewApp(testConfig, storage, zap.L().Sugar())
	r, err := testApp.SetupRouter()
	if err != nil {
		b.Errorf("failed to setup router: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		randURL, _ := utils.GenerateRandomString(length)
		randURL = fmt.Sprintf("%s.ru", randURL)
		reqObj := models.ShortenReq{
			URL: randURL,
		}
		obj, _ := json.Marshal(reqObj)
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(obj))
		req.Header.Add("Content-Type", "application/json")
		b.StartTimer()

		r.ServeHTTP(w, req)

		w.Result()
	}
}

func TestApp_DeleteUserRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = []gin.Param{{Key: auth.UserIDKey, Value: "1"}}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().DeleteMany(gomock.Any(), "1").Return(nil),
	)

	app := NewApp(testConfig, store, zap.L().Sugar())
	r, err := app.SetupRouter()
	if err != nil {
		t.Error(err)
	}

	srv := httptest.NewServer(r)
	defer srv.Close()

	type fields struct {
		config *config.ServerConfig
		store  Store
		logger *zap.SugaredLogger
	}
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test of the test",
			fields: fields{
				config: testConfig,
				store:  store,
				logger: zap.L().Sugar(),
			},
			args: args{
				c: c,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &App{
				config: tt.fields.config,
				store:  tt.fields.store,
				logger: tt.fields.logger,
			}

			tt.args.c.Request = httptest.NewRequest(http.MethodDelete, "/user/urls", bytes.NewBuffer([]byte("[\"1\",\"2\"]")))

			a.DeleteUserRecords(tt.args.c)
		})
	}
}
