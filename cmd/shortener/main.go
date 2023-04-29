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

var urls = make(map[string][]byte)

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

func shortURL(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
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

		return
	} else if req.Method == http.MethodGet {
		reqPathElements := strings.Split(req.URL.Path, "/")
		id := reqPathElements[len(reqPathElements)-1]
		res.Header().Set("Location", string(urls[id]))
		res.WriteHeader(http.StatusTemporaryRedirect)

		return
	}

	res.WriteHeader(http.StatusBadRequest)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, shortURL)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
