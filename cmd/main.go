package main

import (
	"fmt"
	"gate-limiter/internal/app"
	"gate-limiter/internal/metrics"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	configPath := os.Getenv("GATE_LIMITER_CONFIG")
	if configPath == "" {
		configPath = "config.yml"
	}

	// handler
	limitHandler, config, err := app.InitRateLimitHandler() // 초기화가 이루어지는 시점
	if err != nil {
		log.Fatal("Error initializing rate limiter handler", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", limitHandler)

	portString := fmt.Sprintf(":%d", config.RateLimiter.Port)
	err = http.ListenAndServe(portString, metrics.WithMetrics(mux))
	log.Fatal(err)
}
