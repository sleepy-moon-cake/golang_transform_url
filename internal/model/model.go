package model

type ShortenURLRequest struct {
	URL string `json:"url"`
}

type ShortenURLResponse struct {
	Result string `json:"result"`
}

type ShortenURLRecord struct {
	ID          int    `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserUUID    string `json:"user_id"`
	DeletedFlag bool   `db:"is_deleted"`
}

type ShortenUrlDeleteRecord struct {
	UserUUID string `json:"user_id"`
	ShortURL string `json:"short_url"`
}
type ShortenURLBatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenURLBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type URLStats struct {
	Urls  int `json:"urls"`
	Users int `json:"users"`
}
