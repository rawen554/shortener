package main

import (
	"crypto/rand"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
)

const httpProtocol = "http://"

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

func URLHandlerRoot(urls map[string][]byte) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			shortURL(res, req, urls)
			return
		} else if req.Method == http.MethodGet {
			redirectToOriginal(res, req, urls)
			return
		}

		res.WriteHeader(http.StatusBadRequest)
	}
}

func redirectToOriginal(res http.ResponseWriter, req *http.Request, urls map[string][]byte) {
	reqPathElements := strings.Split(req.URL.Path, "/")
	id := reqPathElements[len(reqPathElements)-1]
	originalURL := urls[id]
	if originalURL == nil {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	res.Header().Set("Location", string(originalURL))
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func shortURL(res http.ResponseWriter, req *http.Request, urls map[string][]byte) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		log.Fatal(err)
	}

	id, err := GenerateRandomString(8)
	if err != nil {
		log.Fatal(err)
	}
	urls[id] = body

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(httpProtocol + req.Host + "/" + id))
}

func main() {
	var urls = make(map[string][]byte)
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, URLHandlerRoot(urls))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
