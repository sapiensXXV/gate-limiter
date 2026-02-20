package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"gate-limiter/config/settings"
	"gate-limiter/internal/admin"
	"gate-limiter/internal/app"
	"gate-limiter/internal/buildinfo"
	"gate-limiter/internal/logging"
	"gate-limiter/internal/metrics"
	"gate-limiter/internal/middleware"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var daemon bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the gate-limiter server",
	RunE:  runServer,
}

func init() {
	runCmd.Flags().BoolVarP(&daemon, "daemon", "d", false, "run as background daemon (PID: gl.pid, log: gl.log)")
}

func resolveConfigPath() string {
	if configPath != "" {
		return configPath
	}
	if env := os.Getenv("GATE_LIMITER_CONFIG"); env != "" {
		return env
	}
	return "config.yml"
}

func runServer(_ *cobra.Command, _ []string) error {
	cfgPath := resolveConfigPath()

	config, err := settings.ParseAndValidateConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Daemonize before setting up logging (parent exits immediately)
	if daemon && os.Getenv("GL_DAEMON") != "1" {
		return runDaemon(cfgPath, config)
	}

	closer, err := logging.Setup(config.Logging)
	if err != nil {
		return fmt.Errorf("failed to setup logging: %w", err)
	}
	defer closer()

	// Print banner only in foreground or when output is visible
	if os.Getenv("GL_DAEMON") != "1" {
		settings.PrintBanner(config)
		settings.PrintApiInfo(config.RateLimiter.Apis)
	}

	slog.Info("gate-limiter starting",
		"version", buildinfo.Version,
		"port", config.RateLimiter.Port,
		"admin_port", config.RateLimiter.AdminPort,
		"strategy", config.RateLimiter.Strategy,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startTime := time.Now()

	limitHandler, redisPing, err := app.InitRateLimitHandlerWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to init rate limiter: %w", err)
	}

	server := initMainServer(config, limitHandler, redisPing, startTime)
	adminServer := initAdminServer(config, redisPing, startTime)

	go func() {
		slog.Info("admin server started", "addr", adminServer.Addr)
		if err := adminServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("admin server error", "error", err)
		}
	}()

	go func() {
		<-ctx.Done()
		slog.Info("shutting down servers...")
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("main server shutdown error", "error", err)
		}
		if err := adminServer.Shutdown(context.Background()); err != nil {
			slog.Error("admin server shutdown error", "error", err)
		}
		os.Remove(pidFile)
	}()

	slog.Info("main server started", "addr", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	slog.Info("server stopped")
	return nil
}

func initMainServer(config *settings.RootRateLimiterConfig, handler http.Handler, redisPing func(context.Context) error, startTime time.Time) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		redisStatus := "connected"
		if redisPing != nil {
			if err := redisPing(r.Context()); err != nil {
				redisStatus = "disconnected"
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "UP",
			"uptime":  time.Since(startTime).Truncate(time.Second).String(),
			"redis":   redisStatus,
			"version": buildinfo.Version,
			"apis":    len(config.RateLimiter.Apis),
		})
	})
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/", handler)

	var h http.Handler = mux
	h = metrics.WithMetrics(h)
	h = middleware.WithRequestID(h)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RateLimiter.Port),
		Handler: h,
	}
}

func initAdminServer(config *settings.RootRateLimiterConfig, redisPing func(context.Context) error, startTime time.Time) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", admin.NewStatusHandler(config))
	mux.Handle("/api/status", &admin.StatusAPIHandler{
		Config:    config,
		RedisPing: redisPing,
		StartTime: startTime,
	})

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.RateLimiter.AdminPort),
		Handler: mux,
	}
}

func runDaemon(cfgPath string, config *settings.RootRateLimiterConfig) error {
	args := []string{"run", "-c", cfgPath}
	cmd := exec.Command(os.Args[0], args...)

	if config.Logging.Output == "file" {
		cmd.Stdout = nil
		cmd.Stderr = nil
	} else {
		logFile, err := os.OpenFile("gl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	cmd.Env = append(os.Environ(), "GL_DAEMON=1")

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile("gl.pid", []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write pid file: %w", err)
	}

	fmt.Printf("gate-limiter started (PID: %d)\n", pid)
	fmt.Printf("  server    : http://localhost:%d\n", config.RateLimiter.Port)
	fmt.Printf("  admin page: http://localhost:%d\n", config.RateLimiter.AdminPort)
	if config.Logging.Output == "file" {
		fmt.Printf("  log       : %s/\n", config.Logging.File.Directory)
	} else {
		fmt.Printf("  log       : gl.log\n")
	}
	fmt.Printf("  pid       : gl.pid\n")

	return nil
}
