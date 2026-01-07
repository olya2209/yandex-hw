package models

// Request описывает запрос пользователя.
type Request struct {
	Url string `json:"url"`
}

// Response описывает ответ сервера.
type Response struct {
	ResultURL string `json:"result"`
}
