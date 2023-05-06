package main

import (
	"crypto/rand"
	"io"
	"log"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rawen554/shortener/cmd/config"
)

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func redirectToOriginal(ptrUrls map[string][]byte) func(c *gin.Context) {
	return func(c *gin.Context) {
		res := c.Writer
		id := c.Param("id")
		originalURL := ptrUrls[id]
		if originalURL == nil {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		res.Header().Set("Location", string(originalURL))
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func shortURL(ptrUrls map[string][]byte) func(c *gin.Context) {
	return func(c *gin.Context) {
		req := c.Request
		res := c.Writer
		defer req.Body.Close()
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatal(err)
		}

		id, err := GenerateRandomString(8)
		if err != nil {
			log.Fatal(err)
		}
		ptrUrls[id] = body

		res.Header().Set("Content-Type", "text/plain")
		res.WriteHeader(http.StatusCreated)
		res.Write([]byte(config.Config.RedirectBaseURL + "/" + id))
	}
}

func setupRouter(ptrUrls *map[string][]byte) *gin.Engine {
	r := gin.Default()

	r.GET("/:id", redirectToOriginal(*ptrUrls))
	r.POST("/", shortURL(*ptrUrls))

	return r
}

func main() {
	config.ParseFlags()
	var urls = make(map[string][]byte)

	r := setupRouter(&urls)
	log.Fatal(r.Run(config.Config.FlagRunAddr))
}
