package controllers

import (
	"net/http"

	"github.com/gayratjon-02/Pomosky/internal/platform/httpx"
)

// Health serves GET /healthz for uptime checks and container health checks.
func (c *Controllers) Health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
