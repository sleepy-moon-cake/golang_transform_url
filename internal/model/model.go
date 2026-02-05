package model

type ShortenUrlRequest struct {
	URL string `json:"url"`
}

type ShortenUrlResponse struct {
	Result string `json:"result"`
}
