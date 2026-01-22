package service

import (
	"errors"
	"math/rand"
)

var URLStorage = make(map[string]string)

func CreateshortURL(str string) string {
	key := generateKey()

	URLStorage[key] = str

	return key
}

func GetURLByCode(code string) (string, error) {
	v, ok := URLStorage[code]
	if !ok {
		return "", errors.New("no data")
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
