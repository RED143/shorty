package models

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenBatchRequest []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type ShortenBatchResponse []ShortenBatchResponseItem

type UserURLs struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	IsDeleted   bool   `json:"is_deleted"`
}

type UserURLResponse []UserURLs

type DeleteUrlsRequest []string
