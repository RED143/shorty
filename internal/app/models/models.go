package models

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenBatchRequest []struct {
	CorrelationId int    `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}

type ShortenBatchResponseItem struct {
	CorrelationId int    `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}

type ShortenBatchResponse []ShortenBatchResponseItem
