package httpx

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse 统一的错误响应结构
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// JSON 写出 JSON 响应，设置 Content-Type 与状态码
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error 写出统一错误响应
func Error(w http.ResponseWriter, status int, msg, code string) {
	JSON(w, status, ErrorResponse{Error: msg, Code: code})
}
