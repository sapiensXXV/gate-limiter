package limiter

import (
	"context"
	"embed"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/limiter/store"
	"gate-limiter/internal/limiter/types"
	"gate-limiter/internal/metrics"
	"gate-limiter/internal/middleware"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"time"
)

//go:embed template/default.html
var defaultTemplateFS embed.FS

var defaultPageTemplate = template.Must(template.ParseFS(defaultTemplateFS, "template/default.html"))

const XForwardedFor = "X-Forwarded-For"

type RateLimitHandler struct {
	Limiter     types.RateLimiter
	Proxy       types.ProxyHandler
	Responder   LimitResponder
	Config      settings.RateLimiterConfig
	ClientStore store.CounterStore
	Identifier  ClientIdentifier
}

var _ http.Handler = (*RateLimitHandler)(nil)

func NewRateLimitHandler(
	limiter types.RateLimiter,
	proxy types.ProxyHandler,
	responder LimitResponder,
	config settings.RateLimiterConfig,
	clientStore store.CounterStore,
	identifier ClientIdentifier,
) *RateLimitHandler {
	return &RateLimitHandler{
		Limiter:     limiter,
		Proxy:       proxy,
		Responder:   responder,
		Config:      config,
		ClientStore: clientStore,
		Identifier:  identifier,
	}
}

func (h *RateLimitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := middleware.LoggerFrom(r.Context())

	if h.Config.Target == "" {
		h.respondDefaultPage(w, logger)
		return
	}

	if h.Limiter == nil {
		logger.Error("RateLimitHandler.Limiter is nil")
		return
	}

	policy := h.Config.Strategy
	clientID := h.Identifier.Identify(w, r)

	if exceeded, _, retryAfter := h.isClientLimitExceeded(logger, clientID); exceeded {
		h.Responder.RespondRateLimitExceeded(w, r, 0, retryAfter)
		metrics.ObserveBlocked(policy, "client_limit_exceeded")
		return
	}

	result := h.Limiter.IsTarget(r.Method, r.URL.Path)

	if !result.IsMatch {
		h.Proxy.ToOrigin(w, r, h.Config.Target)
		return
	}

	var queued *types.QueuedRequest
	if h.Config.Strategy == "leaky_bucket" {
		queued = &types.QueuedRequest{
			Request: r,
		}
	}

	decision := h.Limiter.IsAllowed(clientID, result, queued)
	if !decision.Allowed {
		h.Responder.RespondRateLimitExceeded(w, r, decision.Remaining, decision.RetryAfterSec)
		metrics.ObserveBlocked(policy, "허용치 초과")
		return
	}
	metrics.ObserveAllowed(policy)

	h.Proxy.ToOrigin(w, r, h.Config.Target)
}

func (h *RateLimitHandler) isClientLimitExceeded(logger *slog.Logger, clientID string) (exceeded bool, remaining int, retryAfterSec int) {
	if h.Config.Client.Limit <= 0 || h.Config.Client.WindowSeconds <= 0 || h.ClientStore == nil {
		return false, 0, 0
	}

	windowDuration := time.Duration(h.Config.Client.WindowSeconds) * time.Second
	windowStart := time.Now().Truncate(windowDuration)
	key := fmt.Sprintf("client:%s:%d", clientID, windowStart.Unix())

	result, err := h.ClientStore.IncrementAndGet(context.TODO(), key, h.Config.Client.WindowSeconds)
	if err != nil {
		logger.Error("client limit check error", "error", err)
		return false, 0, 0
	}

	if result.Count > int64(h.Config.Client.Limit) {
		retryAt := windowStart.Add(windowDuration)
		wait := time.Until(retryAt)
		sec := int(math.Ceil(wait.Seconds()))
		if sec < 0 {
			sec = 0
		}
		return true, 0, sec
	}

	return false, h.Config.Client.Limit - int(result.Count), 0
}

func (h *RateLimitHandler) respondDefaultPage(w http.ResponseWriter, logger *slog.Logger) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct{ Port int }{Port: h.Config.Port}
	if err := defaultPageTemplate.Execute(w, data); err != nil {
		logger.Error("default page template execute error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
