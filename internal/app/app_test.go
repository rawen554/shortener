package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
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

func BenchmarkShortUrl(b *testing.B) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	length := 10

	storage, err := fs.NewFileStorage(TestStoragePath)
	if err != nil {
		b.Errorf(ErrorSetupStorage, err)
		return
	}
	defer func() {
		if err := storage.DeleteStorageFile(); err != nil {
			b.Errorf(ErrorDeletingTestFile, err)
		}
	}()

	testApp := NewApp(testConfig, storage, zap.L().Sugar())
	r, err := testApp.SetupRouter()
	if err != nil {
		b.Errorf(ErrorSetupRouter, err)
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
		req.Header.Add(contentType, "application/json")
		b.StartTimer()

		r.ServeHTTP(w, req)

		res := w.Result()
		if err := res.Body.Close(); err != nil {
			b.Errorf("cant close body: %v", err)
		}
	}
}

func TestApp_DeleteUserRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().DeleteMany(models.DeleteUserURLsReq{"1", "2"}, gomock.Any()).Return(nil),
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
		url      string
		urls     models.DeleteUserURLsReq
		wantCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "successfull delete",
			fields: fields{
				config: testConfig,
				store:  store,
				logger: zap.L().Sugar(),
			},
			args: args{
				url:      "/api/user/urls",
				urls:     models.DeleteUserURLsReq{"1", "2"},
				wantCode: http.StatusAccepted,
			},
		},
	}
	//nolint: dupl // unnessary deduplication, prevent further expansion
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.args.urls)
			if err != nil {
				t.Error(err)
			}

			url, err := url.JoinPath(srv.URL, tt.args.url)
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(b))
			if err != nil {
				t.Error(err)
			}
			if err := req.Body.Close(); err != nil {
				t.Error(err)
			}

			res, err := srv.Client().Do(req)
			if err != nil {
				t.Error(err)
			}
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.args.wantCode, res.StatusCode)
		})
	}
}

func TestApp_Ping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().Ping().Return(nil),
		store.EXPECT().Ping().Return(fmt.Errorf("lost connection to db")),
	)

	app := NewApp(testConfig, store, zap.L().Sugar())
	r, err := app.SetupRouter()
	if err != nil {
		t.Error(err)
	}

	srv := httptest.NewServer(r)
	defer srv.Close()

	type args struct {
		wantCode int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "connected",
			args: args{
				wantCode: http.StatusOK,
			},
		},
		{
			name: "lost connection",
			args: args{
				wantCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.JoinPath(srv.URL, "/ping")
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Error(err)
			}

			res, err := srv.Client().Do(req)
			if err != nil {
				t.Error(err)
			}
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.args.wantCode, res.StatusCode)
		})
	}
}

func TestApp_ShortenBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().PutBatch(
			[]models.URLBatchReq{
				{
					CorrelationID: "1",
					OriginalURL:   "ya.ru",
				},
			},
			gomock.Any(),
		).Return([]models.URLBatchRes{
			{
				CorrelationID: "1",
				ShortURL:      "http://localhost:8080/abc",
			},
		}, nil),
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
		url      string
		batch    []models.URLBatchReq
		wantCode int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "successfull put batch",
			fields: fields{
				config: testConfig,
				store:  store,
				logger: zap.L().Sugar(),
			},
			args: args{
				url: "/api/shorten/batch",
				batch: []models.URLBatchReq{
					{
						CorrelationID: "1",
						OriginalURL:   "ya.ru",
					},
				},
				wantCode: http.StatusCreated,
			},
		},
	}
	//nolint: dupl // unnessary deduplication
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.args.batch)
			if err != nil {
				t.Error(err)
			}

			url, err := url.JoinPath(srv.URL, tt.args.url)
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
			if err != nil {
				t.Error(err)
			}
			if err := req.Body.Close(); err != nil {
				t.Error(err)
			}

			res, err := srv.Client().Do(req)
			if err != nil {
				t.Error(err)
			}
			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.args.wantCode, res.StatusCode)
		})
	}
}
