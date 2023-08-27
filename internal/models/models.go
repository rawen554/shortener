// Модуль декларирует модели объектов.
package models

// URLRecordFS структура URL записей при работе с файловой системой.
type URLRecordFS struct {
	URLRecord
	UUID   string `json:"uuid"`
	UserID string `json:"user_id"`
}

// URLRecordMemory структура URL записей при работе с памятью.
type URLRecordMemory struct {
	OriginalURL string
	UserID      string
}

// URLRecord ожидаемое тело запроса на сохранение записи URL.
type URLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLBatchReq структура запроса на сохранение батча.
type URLBatchReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// URLBatchRes структура ответа на сохранение батча.
type URLBatchRes struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// DeleteUserURLsReq структура запроса на удаление записей батчем.
type DeleteUserURLsReq []string

// ShortenReq структура запроса на сохранение одного URL.
type ShortenReq struct {
	URL string `json:"url"`
}

// ShortenRes структура ответа на сохранение одного URL.
type ShortenRes struct {
	Result string `json:"result"`
}
