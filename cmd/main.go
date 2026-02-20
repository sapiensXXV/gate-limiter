package main

import (
	"context"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/admin"
	"gate-limiter/internal/app"
	"gate-limiter/internal/metrics"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configPath := os.Getenv("GATE_LIMITER_CONFIG")
	if configPath == "" {
		configPath = "config.yml"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	limitHandler, config, err := app.InitRateLimitHandler(ctx, configPath)
	if err != nil {
		log.Fatal("Error initializing rate limiter handler", err)
	}

	server := initMainServer(config, limitHandler)
	adminServer := initAdminServer(config)

	go func() {
		log.Printf("Admin page: http://localhost:%d\n", config.RateLimiter.AdminPort)
		if err := adminServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("admin server error: %v", err)
		}
	}()

	waitForShutdown(ctx, server, adminServer)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("server stopped")
}

func initMainServer(config *settings.RootRateLimiterConfig, handler http.Handler) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", handler)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RateLimiter.Port),
		Handler: metrics.WithMetrics(mux),
	}
}

func initAdminServer(config *settings.RootRateLimiterConfig) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", admin.NewStatusHandler(config))

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RateLimiter.AdminPort),
		Handler: mux,
	}
}

func waitForShutdown(ctx context.Context, servers ...*http.Server) {
	go func() {
		<-ctx.Done()
		log.Println("shutting down server...")
		for _, s := range servers {
			if err := s.Shutdown(context.Background()); err != nil {
				log.Printf("server shutdown error (%s): %v", s.Addr, err)
			}
		}
	}()
}
