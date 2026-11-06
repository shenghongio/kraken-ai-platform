package v1

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/shenghongio/kraken-ai-platform/platform/health"
)

type HealthHandler struct {
	checker *health.Checker
	logger  *slog.Logger
}

func NewHealthHandler(checker *health.Checker, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{checker: checker, logger: logger}
}

// Healthz  /healthz - 进程存活，不检查依赖
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Readyz /readyz - 就绪，检查DB连通
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	withTimeout, cancelFunc := context.WithTimeout(r.Context(), time.Second*2)
	defer cancelFunc()
	if err := h.checker.Readiness(withTimeout); err != nil {
		h.logger.Warn("readiness check failed", slog.Any("error", err))
		writeError(w, http.StatusServiceUnavailable, "db not ready", "db_unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
