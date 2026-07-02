package handler

// ErrorResponse はHTTP通信において、フロントエンドに一律でエラーを伝えるための器
type ErrorResponse struct {
	Message string `json:"message"`
}
