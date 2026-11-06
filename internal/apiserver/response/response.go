package response

import (
	"net/http"

	"github.com/shenghongio/kraken-ai-platform/platform/httpx"
)

type ErrorResponse = httpx.ErrorResponse

func WriteError(w http.ResponseWriter, status int, msg, code string) {
	httpx.Error(w, status, msg, code)
}
