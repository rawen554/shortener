package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/internal/logic"
	"github.com/rawen554/shortener/internal/models"
	"github.com/rawen554/shortener/internal/store/fs"
	"github.com/rawen554/shortener/internal/utils"
	"go.uber.org/zap"
)

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

	coreLogic := logic.NewCoreLogic(testConfig, storage, zap.L().Sugar())
	testApp := NewApp(testConfig, coreLogic, zap.L().Sugar())
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
