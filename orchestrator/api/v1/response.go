package v1

import (
	"net/http"

	"github.com/shenghongio/kraken-ai-platform/platform/httpx"
)

type ErrorResponse = httpx.ErrorResponse

func writeJSON(w http.ResponseWriter, status int, v any) {
	httpx.JSON(w, status, v)
}

func writeError(w http.ResponseWriter, status int, msg, code string) {
	httpx.Error(w, status, msg, code)
}
