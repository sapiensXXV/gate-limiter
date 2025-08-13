package main

import (
	config_ratelimiter "gate-limiter/config/settings"
	"gate-limiter/internal/app"
	"gate-limiter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func main() {
	configPath := os.Getenv("GATE_LIMITER_CONFIG")
	if configPath == "" {
		configPath = "config.yml"
	}

	_, err := config_ratelimiter.LoadRateLimitConfig(configPath)
	if err != nil {
		log.Fatal("Error loading config.yml file")
	}

	// handler
	limitHandler, err := app.InitRateLimitHandler() // 초기화가 이루어지는 시점
	if err != nil {
		log.Fatal("Error initializing rate limiter handler", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", limitHandler)
	log.Fatal(http.ListenAndServe(":8081", metrics.WithMetrics(mux)))
}
