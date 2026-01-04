package service

import (
	"errors"
	"math/rand"
)

var urlStorage = make(map[string]string)

func CreateShortUrl(str string) string {
	key := generateKey()

	urlStorage[key] = str

	return key
}

func GetUrlByCode(code string) (string, error) {
	v, ok := urlStorage[code]
	if !ok {
		return "", errors.New("No data")
	}
	return v, nil
}

func generateKey() string {
	values := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)

	for i := range b {
		b[i] = values[rand.Intn(len((values)))]
	}
	return string(b)
}
