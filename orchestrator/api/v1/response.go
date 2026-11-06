package v1

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse 统一的错误响应结构，与interanl/apiserver/response 保持一致
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// writeJSON 写出JSON响应
// Content-Type 与状态码行为与 interanl/apiserver/response 完全一致
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// wirteError 写出统一错误响应
func writeError(w http.ResponseWriter, status int, msg, code string) {
	writeJSON(w, status, ErrorResponse{Error: msg, Code: code})
}
