package admin

import (
	"context"
	"encoding/json"
	"gate-limiter/config/settings"
	"gate-limiter/internal/buildinfo"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type StatusAPIHandler struct {
	Config    *settings.RootRateLimiterConfig
	RedisPing func(ctx context.Context) error
	StartTime time.Time
}

type statusAPIResponse struct {
	Uptime       string  `json:"uptime"`
	Version      string  `json:"version"`
	Redis        string  `json:"redis"`
	AllowedTotal float64 `json:"allowed_total"`
	BlockedTotal float64 `json:"blocked_total"`
	ApiCount     int     `json:"api_count"`
}

func (h *StatusAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	redisStatus := "connected"
	if h.RedisPing != nil {
		if err := h.RedisPing(r.Context()); err != nil {
			redisStatus = "disconnected"
		}
	}

	allowed, blocked := getDecisionCounts()

	resp := statusAPIResponse{
		Uptime:       time.Since(h.StartTime).Truncate(time.Second).String(),
		Version:      buildinfo.Version,
		Redis:        redisStatus,
		AllowedTotal: allowed,
		BlockedTotal: blocked,
		ApiCount:     len(h.Config.RateLimiter.Apis),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getDecisionCounts() (allowed, blocked float64) {
	mfs, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		return 0, 0
	}
	for _, mf := range mfs {
		if mf.GetName() != "gatelimiter_rate_limit_decisions_total" {
			continue
		}
		for _, m := range mf.GetMetric() {
			result := ""
			for _, lp := range m.GetLabel() {
				if lp.GetName() == "result" {
					result = lp.GetValue()
				}
			}
			switch result {
			case "allowed":
				allowed += m.GetCounter().GetValue()
			case "blocked":
				blocked += m.GetCounter().GetValue()
			}
		}
	}
	return
}
