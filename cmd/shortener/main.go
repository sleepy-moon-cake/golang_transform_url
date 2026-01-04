package main

import (
	"net/http"
	"os"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/handler"
)

func main() {
	port:= os.Getenv("PORT");
	if port == ""{
		port= "8080"
	}
	url:=":"+port

	mux:= http.NewServeMux()

	mux.HandleFunc("/", handler.Url)

	http.ListenAndServe(url,mux)
}

