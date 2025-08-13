package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var httpReqTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gatelimiter",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total inbound requests",
	},
	[]string{"route", "code", "method"},
)

var httpReqDur = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "gatelimiter",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "End-to-end request latency",
		Buckets:   prometheus.DefBuckets,
	},
	[]string{"route", "code", "method"},
)

var rlDecisionTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "gatelimiter",
		Subsystem: "rate_limit",
		Name:      "decisions_total", // gatelimiter_rate_limit_decisions_total
		Help:      "Rate limit decisions per policy",
	},
	[]string{"policy", "result", "reason"},
)

var rlLimitPerSec = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "gatelimiter",
		Subsystem: "config",
		Name:      "limit_per_sec", // gatelimiter_config_limit_per_sec
		Help:      "Configured per-second limit per policy",
	},
	[]string{"policy"},
)

// 전역 미들웨어: 모든 요청을 계측
func WithMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// /metrics 자체는 계측 제외
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		route := normalizeMetricName(r.URL.Path)

		// route 라벨을 고정(currying)하고, code/method는 promhttp가 채운다.
		h := promhttp.InstrumentHandlerCounter(
			httpReqTotal.MustCurryWith(prometheus.Labels{"route": route}),
			promhttp.InstrumentHandlerDuration(
				httpReqDur.MustCurryWith(prometheus.Labels{"route": route}),
				next,
			),
		)
		h.ServeHTTP(w, r)
	})
}

// 편의 함수. 제한 판단 지점에서 해당 함수를 사용하면 된다.
func ObserveAllowed(policy string) {
	rlDecisionTotal.WithLabelValues(policy, "allowed", "ok").Inc()
}

func ObserveBlocked(policy, reason string) {
	rlDecisionTotal.WithLabelValues(policy, "blocked", reason).Inc()
}

func SetLimitPerSec(policy string, v float64 {
	rlLimitPerSec.WithLabelValues(policy).Set(v)
}
