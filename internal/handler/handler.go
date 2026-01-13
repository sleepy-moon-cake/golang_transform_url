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

	mux.HandleFunc("GET /{shortURL}", getshortURL)
	mux.HandleFunc("POST /", CreateshortURL)
	return mux
}

func CreateshortURL(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortURL := service.CreateshortURL("http://localhost:8080/" + string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func getshortURL(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if r.URL.Path == "/" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")

	originalUrl, err := service.GetUrlByCode(shortURL)

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", "http://localhost:8080/"+originalUrl)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
