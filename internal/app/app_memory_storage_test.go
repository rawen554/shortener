package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rawen554/shortener/internal/config"
	"github.com/rawen554/shortener/internal/middleware/auth"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_DeleteManyInMemory(t *testing.T) {
	type args struct {
		urls       map[string]models.URLRecordMemory
		deleteUrls models.DeleteUserURLsReq
	}
	tests := []struct {
		name          string
		userID        string
		args          args
		wantCode      int
		wantUrlsCount int
	}{
		{
			name:   "delete user url",
			userID: "1",
			args: args{
				urls: map[string]models.URLRecordMemory{
					"1": {OriginalURL: "https://ya.ru", UserID: "1"},
				},
				deleteUrls: models.DeleteUserURLsReq{"1"},
			},
			wantCode:      http.StatusAccepted,
			wantUrlsCount: 0,
		},
		{
			name:   "error deleting another url",
			userID: "2",
			args: args{
				urls: map[string]models.URLRecordMemory{
					"1": {OriginalURL: "https://ya.ru", UserID: "1"},
				},
				deleteUrls: models.DeleteUserURLsReq{"1"},
			},
			wantCode:      http.StatusAccepted,
			wantUrlsCount: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()

			storage, err := memory.NewMemoryStorage(tt.args.urls)
			if err != nil {
				t.Errorf(ErrorSetupStorage, err)
				return
			}

			testApp := NewApp(testConfig, storage, zap.L().Sugar())
			r, err := testApp.SetupRouter()
			if err != nil {
				t.Errorf(ErrorSetupRouter, err)
			}

			obj, err := json.Marshal(tt.args.deleteUrls)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewBuffer(obj))
			req.Header.Add(contentType, applicationJSON)

			token, err := auth.BuildJWTString(testConfig.Secret, tt.userID)
			if err != nil {
				t.Error(err)
			}
			cookie := &http.Cookie{
				Name:   auth.CookieName,
				Value:  token,
				MaxAge: 300,
			}
			req.AddCookie(cookie)

			r.ServeHTTP(w, req)

			res := w.Result()
			if err := res.Body.Close(); err != nil {
				t.Errorf(ErrorClosingBody, err)
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantCode, res.StatusCode)

			timer := time.NewTimer(500 * time.Millisecond)

			<-timer.C

			assert.Equal(t, tt.wantUrlsCount, storage.UrlsCount)
		})
	}
}

func TestApp_GetUserRecordsInMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type args struct {
		url      string
		urls     map[string]models.URLRecordMemory
		userID   string
		wantCode int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successfull get",
			args: args{
				url:      "/api/user/urls",
				urls:     map[string]models.URLRecordMemory{"1": {OriginalURL: "123", UserID: "1"}},
				userID:   "1",
				wantCode: http.StatusOK,
			},
		},
		{
			name: "no content",
			args: args{
				url:      "/api/user/urls",
				urls:     map[string]models.URLRecordMemory{"1": {OriginalURL: "123", UserID: "1"}},
				userID:   "2",
				wantCode: http.StatusNoContent,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			store, err := memory.NewMemoryStorage(tt.args.urls)
			if err != nil {
				t.Errorf(ErrorSetupStorage, err)
				return
			}

			app := NewApp(testConfig, store, zap.L().Sugar())
			r, err := app.SetupRouter()
			if err != nil {
				t.Error(err)
			}

			srv := httptest.NewServer(r)
			defer srv.Close()

			url, err := url.JoinPath(srv.URL, tt.args.url)
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Error(err)
			}

			token, err := auth.BuildJWTString(testConfig.Secret, tt.args.userID)
			if err != nil {
				t.Error(err)
			}
			cookie := &http.Cookie{
				Name:   auth.CookieName,
				Value:  token,
				MaxAge: 300,
			}
			req.AddCookie(cookie)

			res, err := srv.Client().Do(req)
			if err != nil {
				t.Error(err)
			}

			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.args.wantCode, res.StatusCode)
			assert.Equal(t, tt.args.wantCode, res.StatusCode)
		})
	}
}

func TestApp_PingInMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store, err := memory.NewMemoryStorage(make(map[string]models.URLRecordMemory))
	if err != nil {
		t.Errorf(ErrorSetupStorage, err)
		return
	}

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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.JoinPath(srv.URL, pingPath)
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

func TestApp_ShortenBatchInMemory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store, err := memory.NewMemoryStorage(make(map[string]models.URLRecordMemory))
	if err != nil {
		t.Errorf(ErrorSetupStorage, err)
		return
	}

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
