package models

type URLRecordFS struct {
	URLRecord
	UUID   string `json:"uuid"`
	UserID string `json:"user_id"`
}

type URLRecordMemory struct {
	OriginalURL string
	UserID      string
}

type URLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type URLBatchReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
type URLBatchRes struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type DeleteUserURLsReq []string

type ShortenReq struct {
	URL string `json:"url"`
}

type ShortenRes struct {
	Result string `json:"result"`
}
