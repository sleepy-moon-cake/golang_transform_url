package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
)

var URLStorage sync.Map

func CreateShortURL(str string) string {
	key, err := generateKey()
	if err != nil {
		panic(err)
	}

	URLStorage.Store(key, str)

	return key
}

func GetURLByCode(code string) (string, error) {
	v, ok := URLStorage.Load(code)
	if !ok {
		return "", errors.New("no data")
	}

	return v.(string), nil
}

func generateKey() (string, error) {
	const values = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 10

	b := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(values))))
		if err != nil {
			return "", err
		}
		b[i] = values[n.Int64()]
	}

	return string(b), nil
}
