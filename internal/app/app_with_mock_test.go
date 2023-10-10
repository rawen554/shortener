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
	"github.com/rawen554/shortener/internal/logic"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_GetRecord(t *testing.T) {
	const link = "link"

	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().Get("any").Return(link, nil),
	)

	coreLogic := logic.NewCoreLogic(testConfig, store, zap.L().Sugar())
	app := NewApp(testConfig, coreLogic, zap.L().Sugar())
	r, err := app.SetupRouter()
	if err != nil {
		t.Error(err)
	}

	srv := httptest.NewServer(r)
	defer srv.Close()

	type args struct {
		url      string
		response string
		wantCode int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successfull redirect",
			args: args{
				url:      "/any",
				response: "/link",
				wantCode: http.StatusTemporaryRedirect,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.JoinPath(srv.URL, tt.args.url)
			if err != nil {
				t.Error(err)
			}

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				t.Error(err)
			}

			client := srv.Client()
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			res, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}

			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.args.wantCode, res.StatusCode)
			assert.Equal(t, tt.args.response, res.Header.Get("Location"))
		})
	}
}

func Test_PutRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().Put(gomock.Any(), gomock.Any(), gomock.Any()).Return("link", nil),
	)

	coreLogic := logic.NewCoreLogic(testConfig, store, zap.L().Sugar())
	app := NewApp(testConfig, coreLogic, zap.L().Sugar())
	r, err := app.SetupRouter()
	if err != nil {
		t.Error(err)
	}

	srv := httptest.NewServer(r)
	defer srv.Close()

	type args struct {
		url      string
		link     models.ShortenReq
		wantCode int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successfull redirect",
			args: args{
				url:      "/",
				link:     models.ShortenReq{URL: "link"},
				wantCode: http.StatusCreated,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.JoinPath(srv.URL, tt.args.url)
			if err != nil {
				t.Error(err)
			}

			obj, err := json.Marshal(tt.args.link)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(obj))
			if err != nil {
				t.Error(err)
			}

			client := srv.Client()
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}

			res, err := client.Do(req)
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

func TestApp_GetUserRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mocks.NewMockStore(ctrl)

	gomock.InOrder(
		store.EXPECT().GetAllByUserID(gomock.Any()).Return([]models.URLRecord{}, nil),
		store.EXPECT().GetAllByUserID(gomock.Any()).Return([]models.URLRecord{{ShortURL: "test", OriginalURL: "test"}}, nil),
		store.EXPECT().GetAllByUserID(gomock.Any()).Return([]models.URLRecord{}, fmt.Errorf("test error")),
	)

	coreLogic := logic.NewCoreLogic(testConfig, store, zap.L().Sugar())
	app := NewApp(testConfig, coreLogic, zap.L().Sugar())
	r, err := app.SetupRouter()
	if err != nil {
		t.Error(err)
	}

	srv := httptest.NewServer(r)
	defer srv.Close()

	type args struct {
		url      string
		response string
		wantCode int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "no data",
			args: args{
				url:      "/api/user/urls",
				response: "",
				wantCode: http.StatusNoContent,
			},
		},
		{
			name: "successfull get",
			args: args{
				url:      "/api/user/urls",
				response: "[{\"short_url\":\"test\",\"original_url\":\"test\"}]",
				wantCode: http.StatusOK,
			},
		},
		{
			name: "error get",
			args: args{
				url:      "/api/user/urls",
				response: "",
				wantCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			url, err := url.JoinPath(srv.URL, tt.args.url)
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
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Error(err)
			}

			if err := res.Body.Close(); err != nil {
				t.Error(err)
			}

			assert.Equal(t, tt.args.wantCode, res.StatusCode)
			assert.Equal(t, tt.args.response, string(body))
		})
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

	coreLogic := logic.NewCoreLogic(testConfig, store, zap.L().Sugar())
	app := NewApp(testConfig, coreLogic, zap.L().Sugar())
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

	coreLogic := logic.NewCoreLogic(testConfig, store, zap.L().Sugar())
	app := NewApp(testConfig, coreLogic, zap.L().Sugar())
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

	coreLogic := logic.NewCoreLogic(testConfig, store, zap.L().Sugar())
	app := NewApp(testConfig, coreLogic, zap.L().Sugar())
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
