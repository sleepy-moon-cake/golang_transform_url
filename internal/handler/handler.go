package handler

import (
	"io"
	"net/http"
	"strings"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
)

func ListenAndServe(cng *config.Config) error {
	handler := createHandler()
	return http.ListenAndServe(cng.ServerAddress, handler)
}

func createHandler() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{shortUrl}", getShortUrl)
	mux.HandleFunc("POST /", createShortUrl)
	return mux
}

func createShortUrl(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortUrl := service.CreateShortUrl(string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortUrl))
}

func getShortUrl(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if r.URL.Path == "/" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortUrl := strings.TrimPrefix(r.URL.Path, "/")

	originalUrl, err := service.GetUrlByCode(shortUrl)

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", originalUrl)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
