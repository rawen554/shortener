package utils

import (
	"errors"
	"math/rand"
)

func GenerateRandomString(n int) (string, error) {
	if n < 0 {
		return "", errors.New("negative length")
	}
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num := rand.Intn(len(letters))
		ret[i] = letters[num]
	}

	return string(ret), nil
}
