package main

import (
	"context"
	"fmt"
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

	// handler
	limitHandler, config, err := app.InitRateLimitHandler(ctx, configPath) // 초기화가 이루어지는 시점
	if err != nil {
		log.Fatal("Error initializing rate limiter handler", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", limitHandler)

	portString := fmt.Sprintf(":%d", config.RateLimiter.Port)
	server := &http.Server{
		Addr:    portString,
		Handler: metrics.WithMetrics(mux),
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("server shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
